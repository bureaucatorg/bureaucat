-- Widen color columns to support CSS color values from Plane imports
ALTER TABLE project_states ALTER COLUMN color TYPE VARCHAR(50);
ALTER TABLE project_labels ALTER COLUMN color TYPE VARCHAR(50);
