CREATE TABLE ad_campaigns (
    id bigserial PRIMARY KEY,
    title varchar(128) NOT NULL CHECK (title <> ''),
    message text NOT NULL CHECK (message <> '' AND octet_length(message) <= 1024),
    entities jsonb NOT NULL DEFAULT '[]'::jsonb,
    button_text varchar(64) NOT NULL DEFAULT '',
    url varchar(512) NOT NULL CHECK (url <> ''),
    sponsor_info varchar(128) NOT NULL DEFAULT '',
    active boolean NOT NULL DEFAULT true,
    target_channel_ids bigint[] NOT NULL DEFAULT '{}',
    impression_count bigint NOT NULL DEFAULT 0 CHECK (impression_count >= 0),
    view_count bigint NOT NULL DEFAULT 0 CHECK (view_count >= 0),
    click_count bigint NOT NULL DEFAULT 0 CHECK (click_count >= 0),
    created_by varchar(128) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ad_campaigns_active_idx ON ad_campaigns (id) WHERE active;
CREATE INDEX ad_campaigns_target_channel_ids_idx ON ad_campaigns USING gin (target_channel_ids)
    WHERE active AND target_channel_ids <> '{}';
