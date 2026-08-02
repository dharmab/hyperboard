/*
 * frame — extract a single video frame as raw RGBA.
 *
 * Compiled to wasm32-wasip1 and linked against a decode-only, LGPL-only build
 * of FFmpeg. Run under wazero by internal/media/ffwasm.
 *
 * Usage: frame <input-path> <offset-seconds>
 *
 * Writes to stdout:
 *   magic  "FRM1"      4 bytes
 *   width  uint32 LE   4 bytes
 *   height uint32 LE   4 bytes
 *   pixels width*height*4 bytes, RGBA
 *
 * Diagnostics go to stderr. Exit status is 0 on success.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/imgutils.h>
#include <libswscale/swscale.h>

static int write_frame(const AVFrame *frame) {
    int width = frame->width;
    int height = frame->height;

    struct SwsContext *sws = sws_getContext(
        width, height, (enum AVPixelFormat)frame->format,
        width, height, AV_PIX_FMT_RGBA,
        SWS_BILINEAR, NULL, NULL, NULL);
    if (!sws) {
        fprintf(stderr, "frame: cannot init swscale for pix_fmt %d\n", frame->format);
        return 1;
    }

    uint8_t *dst_data[4] = {NULL};
    int dst_linesize[4] = {0};
    int ret = av_image_alloc(dst_data, dst_linesize, width, height, AV_PIX_FMT_RGBA, 1);
    if (ret < 0) {
        fprintf(stderr, "frame: cannot allocate %dx%d rgba buffer\n", width, height);
        sws_freeContext(sws);
        return 1;
    }

    sws_scale(sws, (const uint8_t *const *)frame->data, frame->linesize, 0, height,
              dst_data, dst_linesize);
    sws_freeContext(sws);

    unsigned char header[12] = {'F', 'R', 'M', '1'};
    uint32_t w = (uint32_t)width, h = (uint32_t)height;
    header[4] = w & 0xff;
    header[5] = (w >> 8) & 0xff;
    header[6] = (w >> 16) & 0xff;
    header[7] = (w >> 24) & 0xff;
    header[8] = h & 0xff;
    header[9] = (h >> 8) & 0xff;
    header[10] = (h >> 16) & 0xff;
    header[11] = (h >> 24) & 0xff;

    int rc = 0;
    if (fwrite(header, 1, sizeof(header), stdout) != sizeof(header)) {
        rc = 1;
    }
    for (int y = 0; rc == 0 && y < height; y++) {
        const uint8_t *row = dst_data[0] + (size_t)y * dst_linesize[0];
        if (fwrite(row, 1, (size_t)width * 4, stdout) != (size_t)width * 4) {
            rc = 1;
        }
    }
    if (fflush(stdout) != 0) {
        rc = 1;
    }
    if (rc != 0) {
        fprintf(stderr, "frame: short write to stdout\n");
    }

    av_freep(&dst_data[0]);
    return rc;
}

int main(int argc, char **argv) {
    if (argc != 3) {
        fprintf(stderr, "usage: frame <input-path> <offset-seconds>\n");
        return 2;
    }
    const char *path = argv[1];
    double offset = strtod(argv[2], NULL);
    if (offset < 0) {
        offset = 0;
    }

    av_log_set_level(AV_LOG_ERROR);

    AVFormatContext *fmt = NULL;
    int ret = avformat_open_input(&fmt, path, NULL, NULL);
    if (ret < 0) {
        fprintf(stderr, "frame: open input: %s\n", av_err2str(ret));
        return 1;
    }
    ret = avformat_find_stream_info(fmt, NULL);
    if (ret < 0) {
        fprintf(stderr, "frame: find stream info: %s\n", av_err2str(ret));
        avformat_close_input(&fmt);
        return 1;
    }

    const AVCodec *codec = NULL;
    int stream_index = av_find_best_stream(fmt, AVMEDIA_TYPE_VIDEO, -1, -1, &codec, 0);
    if (stream_index < 0) {
        fprintf(stderr, "frame: no video stream: %s\n", av_err2str(stream_index));
        avformat_close_input(&fmt);
        return 1;
    }
    AVStream *stream = fmt->streams[stream_index];

    AVCodecContext *dec = avcodec_alloc_context3(codec);
    if (!dec) {
        fprintf(stderr, "frame: cannot allocate decoder\n");
        avformat_close_input(&fmt);
        return 1;
    }
    ret = avcodec_parameters_to_context(dec, stream->codecpar);
    if (ret < 0) {
        fprintf(stderr, "frame: codec parameters: %s\n", av_err2str(ret));
        goto fail;
    }
    dec->thread_count = 1;
    ret = avcodec_open2(dec, codec, NULL);
    if (ret < 0) {
        fprintf(stderr, "frame: open decoder %s: %s\n", codec->name, av_err2str(ret));
        goto fail;
    }

    /* Seek to the keyframe at or before the requested offset. */
    int64_t target = (int64_t)(offset / av_q2d(stream->time_base));
    if (stream->start_time != AV_NOPTS_VALUE) {
        target += stream->start_time;
    }
    if (offset > 0) {
        ret = av_seek_frame(fmt, stream_index, target, AVSEEK_FLAG_BACKWARD);
        if (ret < 0) {
            /* Not fatal — fall back to decoding from the start. */
            fprintf(stderr, "frame: seek failed, decoding from start: %s\n", av_err2str(ret));
        }
        avcodec_flush_buffers(dec);
    }

    AVPacket *pkt = av_packet_alloc();
    AVFrame *frame = av_frame_alloc();
    AVFrame *best = av_frame_alloc();
    if (!pkt || !frame || !best) {
        fprintf(stderr, "frame: cannot allocate packet or frame\n");
        ret = 1;
        goto cleanup;
    }

    int have_best = 0;
    int done = 0;
    int draining = 0;

    while (!done) {
        if (!draining) {
            ret = av_read_frame(fmt, pkt);
            if (ret == AVERROR_EOF) {
                draining = 1;
                avcodec_send_packet(dec, NULL);
            } else if (ret < 0) {
                fprintf(stderr, "frame: read packet: %s\n", av_err2str(ret));
                break;
            } else if (pkt->stream_index != stream_index) {
                av_packet_unref(pkt);
                continue;
            } else {
                ret = avcodec_send_packet(dec, pkt);
                av_packet_unref(pkt);
                if (ret < 0 && ret != AVERROR(EAGAIN)) {
                    fprintf(stderr, "frame: send packet: %s\n", av_err2str(ret));
                    break;
                }
            }
        }

        for (;;) {
            ret = avcodec_receive_frame(dec, frame);
            if (ret == AVERROR(EAGAIN)) {
                break;
            }
            if (ret == AVERROR_EOF) {
                done = 1;
                break;
            }
            if (ret < 0) {
                fprintf(stderr, "frame: receive frame: %s\n", av_err2str(ret));
                done = 1;
                break;
            }

            av_frame_unref(best);
            av_frame_move_ref(best, frame);
            have_best = 1;

            int64_t pts = best->best_effort_timestamp;
            if (pts == AV_NOPTS_VALUE) {
                pts = best->pts;
            }
            /* Keep decoding until we reach the requested offset. */
            if (pts == AV_NOPTS_VALUE || pts >= target) {
                done = 1;
                break;
            }
        }
    }

    if (!have_best) {
        fprintf(stderr, "frame: no frame decoded\n");
        ret = 1;
        goto cleanup;
    }

    ret = write_frame(best);

cleanup:
    av_frame_free(&best);
    av_frame_free(&frame);
    av_packet_free(&pkt);
fail:
    avcodec_free_context(&dec);
    avformat_close_input(&fmt);
    return ret == 0 ? 0 : 1;
}
