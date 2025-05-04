
-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;



CREATE TABLE IF NOT EXISTS room_memberships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id VARCHAR(255) NOT NULL,  -- Can be UUID or string ID
  room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (user_id, room_id)
);

CREATE INDEX idx_room_memberships_user_id ON room_memberships(user_id);
CREATE INDEX idx_room_memberships_room_id ON room_memberships(room_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE room_memberships;
-- +goose StatementEnd
