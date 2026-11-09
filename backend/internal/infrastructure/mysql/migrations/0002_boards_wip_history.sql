CREATE TABLE boards (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(120) NOT NULL,
  description TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) CHARACTER SET utf8mb4;

INSERT INTO boards (name, description) VALUES ('Principal', 'Tablero inicial');

ALTER TABLE columns_
  ADD COLUMN board_id INT NULL,
  ADD COLUMN wip_limit INT NULL,
  ADD COLUMN kind ENUM('normal','active','done') NOT NULL DEFAULT 'normal';

UPDATE columns_ SET board_id = (SELECT MIN(id) FROM boards);
UPDATE columns_ SET kind = 'active' WHERE title = 'En progreso';
UPDATE columns_ SET kind = 'done' WHERE title = 'Hecho';

ALTER TABLE columns_
  MODIFY board_id INT NOT NULL,
  ADD CONSTRAINT fk_columns_board FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
  ADD INDEX idx_columns_board (board_id, position);

ALTER TABLE cards
  ADD COLUMN priority ENUM('low','medium','high','urgent') NOT NULL DEFAULT 'medium';

CREATE TABLE card_events (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  card_id INT NOT NULL,
  from_column_id INT NULL,
  to_column_id INT NULL,
  to_kind ENUM('normal','active','done') NOT NULL DEFAULT 'normal',
  moved_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  user_id INT NULL,
  CONSTRAINT fk_events_card FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE,
  CONSTRAINT fk_events_from FOREIGN KEY (from_column_id) REFERENCES columns_(id) ON DELETE SET NULL,
  CONSTRAINT fk_events_to FOREIGN KEY (to_column_id) REFERENCES columns_(id) ON DELETE SET NULL,
  INDEX idx_events_card (card_id, moved_at)
) CHARACTER SET utf8mb4;

INSERT INTO card_events (card_id, to_column_id, to_kind, moved_at)
SELECT c.id, c.column_id, col.kind, c.created_at FROM cards c JOIN columns_ col ON col.id = c.column_id;
