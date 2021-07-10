CREATE OR REPLACE FUNCTION update_modified_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.ModifiedAt = now();
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';