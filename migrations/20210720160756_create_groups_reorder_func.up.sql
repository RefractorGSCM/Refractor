DO $$ BEGIN
    CREATE TYPE reorder_groups_info AS (
        GroupID INT,
        NewPos INT
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE OR REPLACE FUNCTION reorder_groups(newpos reorder_groups_info[]) RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE ginfo reorder_groups_info;
BEGIN
    FOREACH ginfo IN ARRAY newpos
    LOOP
        UPDATE Groups SET Position = ginfo.NewPos WHERE GroupID = ginfo.GroupID;
    END LOOP;
END; $$;