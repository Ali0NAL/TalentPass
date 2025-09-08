-- events.user_id ve application_id zorunlu, FK'ler eklensin

-- 1) (opsiyonel) NULL kayıt var mı?
-- SELECT count(*) FROM events WHERE user_id IS NULL OR application_id IS NULL;

-- 2) NULL varsa, ya sil ya da set et:
-- DELETE FROM events WHERE user_id IS NULL OR application_id IS NULL;
-- veya:
-- UPDATE events SET user_id = 1 WHERE user_id IS NULL;
-- UPDATE events SET application_id = 1 WHERE application_id IS NULL;

-- 3) NOT NULL yap
ALTER TABLE events
  ALTER COLUMN user_id SET NOT NULL,
  ALTER COLUMN application_id SET NOT NULL;

-- 4) FK ekle (varsa eski constraint isimleri farklıysa önce DROP gerekebilir)
ALTER TABLE events
  ADD CONSTRAINT events_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  ADD CONSTRAINT events_application_id_fkey
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;
