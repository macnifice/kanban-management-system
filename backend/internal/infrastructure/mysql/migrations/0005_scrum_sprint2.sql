-- Evidencia 3 (Simulación de un Sprint): Sprint 2 del 21 al 25 de noviembre de 2026.
-- Retoma las 8 tareas que el Sprint 1 dejó sin terminar y carga el historial de movimientos de la simulación
-- (movimientos fechados del 22 al 24 de noviembre). Horas en UTC (México = UTC-6).

INSERT INTO boards (name, description, created_at)
VALUES ('Sprint 2', 'Sprint del 21 al 25 de noviembre de 2026. Objetivo: completar HU-02, HU-03 y HU-04: columnas con límite de WIP editable, detalle de tarjeta y movimiento de tarjetas que respete el WIP, incluso con peticiones simultáneas.', '2026-11-21 23:00:00');
SET @sb := LAST_INSERT_ID();

INSERT INTO columns_ (board_id, title, position, wip_limit, kind) VALUES (@sb, 'Por hacer', 0, NULL, 'normal');
SET @c_todo := LAST_INSERT_ID();
INSERT INTO columns_ (board_id, title, position, wip_limit, kind) VALUES (@sb, 'En proceso', 1, 3, 'active');
SET @c_doing := LAST_INSERT_ID();
INSERT INTO columns_ (board_id, title, position, wip_limit, kind) VALUES (@sb, 'Hecho', 2, NULL, 'done');
SET @c_done := LAST_INSERT_ID();

INSERT INTO labels (board_id, name, color) VALUES (@sb, 'HU-02', '#e08a1e');
SET @l2 := LAST_INSERT_ID();
INSERT INTO labels (board_id, name, color) VALUES (@sb, 'HU-03', '#8250df');
SET @l3 := LAST_INSERT_ID();
INSERT INTO labels (board_id, name, color) VALUES (@sb, 'HU-04', '#d64545');
SET @l4 := LAST_INSERT_ID();
INSERT INTO labels (board_id, name, color) VALUES (@sb, 'Arrastrada del Sprint 1', '#6e7781');
SET @lc := LAST_INSERT_ID();

SET @u := (SELECT MIN(id) FROM users);

-- Posiciones finales: Hecho en orden de término, En proceso la tarea que no se terminó.
INSERT INTO cards (column_id, title, description, position, priority, assignee_id, created_at) VALUES
 (@c_done, 'HU-02 · T2 Validar el límite de WIP', 'La API rechaza límites menores a 1; vacío significa sin límite. Estaba En proceso en el Sprint 1.', 0, 'high', @u, '2026-11-21 23:01:00');
SET @t22 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, assignee_id, created_at) VALUES
 (@c_done, 'HU-02 · T3 Panel para editar nombre y límite', 'Botón ⚙ en cada columna para cambiar su nombre, límite de WIP y tipo. Estaba Por hacer en el Sprint 1.', 4, 'medium', @u, '2026-11-21 23:02:00');
SET @t23 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, assignee_id, created_at) VALUES
 (@c_done, 'HU-03 · T2 Formulario rápido en cada columna', 'Agregar una tarjeta desde el pie de la columna con solo escribir el título. Estaba En proceso en el Sprint 1.', 3, 'medium', @u, '2026-11-21 23:03:00');
SET @t32 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, assignee_id, created_at) VALUES
 (@c_done, 'HU-03 · T3 Modal de detalle de la tarjeta', 'Ver y editar título, descripción y prioridad; mostrar la fecha de creación. Estaba Por hacer en el Sprint 1.', 5, 'medium', @u, '2026-11-21 23:04:00');
SET @t33 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, assignee_id, created_at) VALUES
 (@c_done, 'HU-04 · T1 Endpoint para mover tarjetas', 'PUT /cards/:id/move con columna y posición destino, reordenando en una transacción. Estaba En proceso en el Sprint 1.', 1, 'high', @u, '2026-11-21 23:05:00');
SET @t41 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, assignee_id, created_at) VALUES
 (@c_done, 'HU-04 · T2 Regla de WIP en el dominio', 'Responder 409 wip_limit al mover o crear en una columna llena. Estaba Por hacer en el Sprint 1.', 2, 'high', @u, '2026-11-21 23:06:00');
SET @t42 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, assignee_id, created_at) VALUES
 (@c_done, 'HU-04 · T3 WIP correcto con peticiones simultáneas', 'Bloquear las columnas en la transacción para que el límite se respete aunque lleguen varios movimientos a la vez. Estaba Por hacer en el Sprint 1.', 6, 'high', @u, '2026-11-21 23:07:00');
SET @t43 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, assignee_id, created_at) VALUES
 (@c_doing, 'HU-04 · T4 Arrastrar tarjetas con aviso de columna llena', 'Drag & drop entre columnas; si la columna está llena se revierte y se muestra el aviso. Estaba Por hacer en el Sprint 1.', 0, 'medium', @u, '2026-11-21 23:08:00');
SET @t44 := LAST_INSERT_ID();

INSERT INTO card_labels (card_id, label_id) VALUES
 (@t22, @l2), (@t23, @l2), (@t32, @l3), (@t33, @l3), (@t41, @l4), (@t42, @l4), (@t43, @l4), (@t44, @l4),
 (@t22, @lc), (@t23, @lc), (@t32, @lc), (@t33, @lc), (@t41, @lc), (@t42, @lc), (@t43, @lc), (@t44, @lc);

-- Todas nacen en Por hacer en el Sprint Planning.
INSERT INTO card_events (card_id, from_column_id, to_column_id, to_kind, moved_at)
SELECT c.id, NULL, @c_todo, 'normal', c.created_at FROM cards c WHERE c.id BETWEEN @t22 AND @t44;

-- Movimientos simulados (hora UTC; México = UTC-6). El WIP de En proceso nunca pasa de 3.
INSERT INTO card_events (card_id, from_column_id, to_column_id, to_kind, moved_at) VALUES
 -- Domingo 22
 (@t41, @c_todo,  @c_doing, 'active', '2026-11-22 15:30:00'),
 (@t22, @c_todo,  @c_doing, 'active', '2026-11-22 15:35:00'),
 (@t22, @c_doing, @c_done,  'done',   '2026-11-22 19:00:00'),
 (@t41, @c_doing, @c_done,  'done',   '2026-11-22 21:30:00'),
 (@t32, @c_todo,  @c_doing, 'active', '2026-11-22 21:40:00'),
 -- Lunes 23
 (@t42, @c_todo,  @c_doing, 'active', '2026-11-23 15:20:00'),
 (@t23, @c_todo,  @c_doing, 'active', '2026-11-23 17:10:00'),
 (@t42, @c_doing, @c_done,  'done',   '2026-11-23 22:00:00'),
 (@t43, @c_todo,  @c_doing, 'active', '2026-11-23 22:15:00'),
 (@t32, @c_doing, @c_done,  'done',   '2026-11-23 22:30:00'),
 -- Martes 24
 (@t33, @c_todo,  @c_doing, 'active', '2026-11-24 15:30:00'),
 (@t23, @c_doing, @c_done,  'done',   '2026-11-24 18:00:00'),
 (@t44, @c_todo,  @c_doing, 'active', '2026-11-24 18:10:00'),
 (@t33, @c_doing, @c_done,  'done',   '2026-11-24 21:00:00'),
 (@t43, @c_doing, @c_done,  'done',   '2026-11-25 00:30:00');
