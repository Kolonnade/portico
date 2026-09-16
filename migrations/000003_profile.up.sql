-- A profile the account holder chooses: a display name and an emoji avatar.
-- The avatar is a key from internal/profile, validated there, not an emoji.
ALTER TABLE users ADD COLUMN avatar TEXT;

-- Registration used to copy the email address into display_name, which put the
-- address into every passkey's user name and onto the home page. Clear those
-- copies; the holder sets a real name from the profile page.
UPDATE users u
   SET display_name = NULL
 WHERE EXISTS (SELECT 1 FROM identifiers i
                WHERE i.user_id = u.id
                  AND i.kind = 'email'
                  AND i.value = u.display_name);
