CREATE TABLE tag_cascades (
    tag_id         UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    cascaded_tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (tag_id, cascaded_tag_id),
    CONSTRAINT chk_no_self_cascade CHECK (tag_id != cascaded_tag_id)
);
CREATE INDEX idx_tag_cascades_cascaded_tag_id ON tag_cascades(cascaded_tag_id);
