-- Enforce patient business identity as (name, mobile) while keeping id as PK.

DO $$
DECLARE
  has_patients boolean;
  has_phone boolean;
BEGIN
  SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name = 'patients'
  ) INTO has_patients;

  IF NOT has_patients THEN
    RETURN;
  END IF;

  ALTER TABLE public.patients
    ADD COLUMN IF NOT EXISTS mobile VARCHAR(50);

  SELECT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'patients'
      AND column_name = 'phone'
  ) INTO has_phone;

  IF has_phone THEN
    UPDATE public.patients
    SET mobile = COALESCE(NULLIF(mobile, ''), NULLIF(phone, ''), 'UNKNOWN-' || id::text)
    WHERE mobile IS NULL OR mobile = '';
  ELSE
    UPDATE public.patients
    SET mobile = COALESCE(NULLIF(mobile, ''), 'UNKNOWN-' || id::text)
    WHERE mobile IS NULL OR mobile = '';
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name = 'patients'
  ) THEN
    ALTER TABLE public.patients
      ALTER COLUMN mobile SET NOT NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'patients_name_mobile_key'
      AND conrelid = (
        SELECT oid FROM pg_class
        WHERE relnamespace = 'public'::regnamespace
          AND relname = 'patients'
      )
  ) THEN
    IF EXISTS (
      SELECT 1
      FROM information_schema.tables
      WHERE table_schema = 'public'
        AND table_name = 'patients'
    ) THEN
      ALTER TABLE public.patients
        ADD CONSTRAINT patients_name_mobile_key UNIQUE (name, mobile);
    END IF;
  END IF;
END $$;

DROP INDEX IF EXISTS public.idx_patients_phone;
