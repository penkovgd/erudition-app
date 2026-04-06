ALTER TABLE questions
ADD COLUMN IF NOT EXISTS q_type VARCHAR(50) DEFAULT 'MCQ';
ALTER TABLE questions
ADD COLUMN IF NOT EXISTS payload JSONB;
UPDATE questions
SET payload = jsonb_build_object(
        'options',
        options,
        'correct_key',
        correct_key,
        'image_url',
        image_url
    )
WHERE payload IS NULL;
ALTER TABLE questions
ALTER COLUMN payload
SET NOT NULL;
ALTER TABLE questions DROP COLUMN options;
ALTER TABLE questions DROP COLUMN correct_key;
ALTER TABLE questions DROP COLUMN image_url;    