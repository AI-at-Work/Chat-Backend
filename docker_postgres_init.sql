DO $$
BEGIN

    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    -- Create User_Data table if not exists
    IF NOT EXISTS (SELECT * FROM information_schema.tables WHERE table_name = 'user_data') THEN
        CREATE TABLE User_Data (
           User_Id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
           UserName VARCHAR(255) NOT NULL,
           Models INT[],
           Created_At TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
    END IF;

    -- Create Model_Details table if not exists
    IF NOT EXISTS (SELECT * FROM information_schema.tables WHERE table_name = 'model_details') THEN
        CREATE TABLE Model_Details (
            Model_Id INT PRIMARY KEY,
            Model_Name VARCHAR(255) NOT NULL,
            context_length INT NOT NULL
        );
    END IF;

    -- Create Session_Details table if not exists
    IF NOT EXISTS (SELECT * FROM information_schema.tables WHERE table_name = 'session_details') THEN
        CREATE TABLE Session_Details (
            Session_Id UUID PRIMARY KEY,
            Session_Name VARCHAR(255) NOT NULL,
            User_Id UUID NOT NULL,
            Model_Id INT NOT NULL,
            Created_At TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (User_Id) REFERENCES User_Data(User_Id) ON DELETE CASCADE,
            FOREIGN KEY (Model_Id) REFERENCES Model_Details(Model_Id)
        );
    END IF;

    -- Create User_Data table if not exists
    IF NOT EXISTS (SELECT * FROM information_schema.tables WHERE table_name = 'file_data') THEN
        CREATE TABLE File_Data (
           Session_Id UUID NOT NULL,
           File_Name VARCHAR(255) NOT NULL,
           Created_At TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
           FOREIGN KEY (Session_Id) REFERENCES Session_Details(Session_Id) ON DELETE CASCADE
        );
    END IF;

    -- Create Chat_Details table if not exists
    IF NOT EXISTS (SELECT * FROM information_schema.tables WHERE table_name = 'chat_details') THEN
        CREATE TABLE Chat_Details (
            Session_Id UUID PRIMARY KEY,
            Session_Prompt TEXT NOT NULL,
            Chats_Vector JSONB NOT NULL DEFAULT '[]'::JSONB,
            Chats JSONB NOT NULL DEFAULT '[]'::JSONB,
            FOREIGN KEY (Session_Id) REFERENCES Session_Details(Session_Id) ON DELETE CASCADE
        );
    END IF;

    -- Create trigger function if it doesn't already exist
    IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'append_chat_jsonb') THEN
        CREATE OR REPLACE FUNCTION append_chat_jsonb() RETURNS TRIGGER AS $emp_audit$
        BEGIN
            --
            -- Check if operation is UPDATE and there's new data to add
            IF TG_OP = 'UPDATE' AND NEW.Chats IS NOT NULL THEN
                -- Append new chat entries to existing JSONB array
                NEW.Chats := OLD.Chats || NEW.Chats;
            END IF;

            IF TG_OP = 'UPDATE' AND NEW.Chats_Vector IS NOT NULL THEN
                -- Append new chat entries to existing JSONB array
                NEW.Chats_Vector := OLD.Chats_Vector || NEW.Chats_Vector;
            END IF;
            RETURN NEW;
        END;
        $emp_audit$ LANGUAGE plpgsql;
    END IF;

    -- Create trigger if it doesn't already exist
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_append_chat_jsonb') THEN
        CREATE TRIGGER trg_append_chat_jsonb
            BEFORE UPDATE ON Chat_Details
            FOR EACH ROW
        EXECUTE FUNCTION append_chat_jsonb();
    END IF;

END$$;

COMMIT;

