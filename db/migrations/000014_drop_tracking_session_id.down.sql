ALTER TABLE pasos_portico
    ADD COLUMN IF NOT EXISTS tracking_session_id UUID NULL;
