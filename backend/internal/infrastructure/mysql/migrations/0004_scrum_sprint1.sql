-- Datos de la actividad Scrum (Tema 4): Product Backlog con HU-01..HU-10 y tablero del Sprint 1.
-- Las horas están en UTC (20-21 nov 2026, hora de México = UTC-6).

-- ============ Product Backlog ============
INSERT INTO boards (name, description, created_at)
VALUES ('Product Backlog', 'Historias de usuario HU-01 a HU-10 del Dashboard Kanban, ordenadas por prioridad.', '2026-11-20 16:00:00');
SET @pb := LAST_INSERT_ID();

INSERT INTO columns_ (board_id, title, position, wip_limit, kind) VALUES (@pb, 'Prioridad alta', 0, NULL, 'normal');
SET @pb_alta := LAST_INSERT_ID();
INSERT INTO columns_ (board_id, title, position, wip_limit, kind) VALUES (@pb, 'Prioridad media', 1, NULL, 'normal');
SET @pb_media := LAST_INSERT_ID();
INSERT INTO columns_ (board_id, title, position, wip_limit, kind) VALUES (@pb, 'Prioridad baja', 2, NULL, 'normal');
SET @pb_baja := LAST_INSERT_ID();

INSERT INTO labels (board_id, name, color) VALUES (@pb, 'Requisito del PO', '#2f6fed');
SET @l_po := LAST_INSERT_ID();
INSERT INTO labels (board_id, name, color) VALUES (@pb, 'Propuesta del equipo', '#8250df');
SET @l_eq := LAST_INSERT_ID();
INSERT INTO labels (board_id, name, color) VALUES (@pb, 'Sprint 1', '#1a7f37');
SET @l_s1 := LAST_INSERT_ID();

INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_alta, 'HU-01 · Gestionar tableros', 'Como usuario, quiero crear, editar y eliminar tableros, para organizar cada proyecto en su propio espacio.', 0, 'high', '2026-11-20 16:05:00');
SET @hu01 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_alta, 'HU-02 · Columnas con límite de WIP', 'Como usuario, quiero definir columnas con nombre, posición y límite de WIP, para representar el ciclo de vida de mi trabajo.', 1, 'high', '2026-11-20 16:06:00');
SET @hu02 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_alta, 'HU-03 · Registrar tarjetas', 'Como usuario, quiero crear tarjetas con título, descripción, prioridad y fecha de creación, para registrar las tareas de mi proyecto.', 2, 'high', '2026-11-20 16:07:00');
SET @hu03 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_alta, 'HU-04 · Mover tarjetas respetando el WIP', 'Como usuario, quiero mover tarjetas entre columnas y que el sistema bloquee el movimiento cuando la columna destino llegó a su límite de WIP, para no acumular más trabajo del que el equipo puede atender.', 3, 'high', '2026-11-20 16:08:00');
SET @hu04 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_media, 'HU-05 · Dashboard de métricas', 'Como líder del proyecto, quiero ver un dashboard con las tarjetas por columna, el WIP actual contra su límite y el cycle time, para detectar cuellos de botella.', 0, 'medium', '2026-11-20 16:09:00');
SET @hu05 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_media, 'HU-06 · Buscar y filtrar tarjetas', 'Como usuario, quiero buscar tarjetas por texto y filtrarlas por prioridad o etiqueta, para encontrar rápido una tarea en tableros grandes.', 1, 'medium', '2026-11-20 16:10:00');
SET @hu06 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_media, 'HU-07 · Fechas límite', 'Como usuario, quiero asignar una fecha límite a cada tarjeta y ver resaltadas las vencidas, para atender primero lo que está por vencer.', 2, 'medium', '2026-11-20 16:11:00');
SET @hu07 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_baja, 'HU-08 · Comentarios en tarjetas', 'Como integrante del equipo, quiero comentar dentro de una tarjeta, para resolver dudas sobre la tarea sin salir del tablero.', 0, 'low', '2026-11-20 16:12:00');
SET @hu08 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_baja, 'HU-09 · Exportar métricas a CSV', 'Como líder del proyecto, quiero exportar las métricas del tablero a un archivo CSV, para incluirlas en los reportes de avance.', 1, 'low', '2026-11-20 16:13:00');
SET @hu09 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@pb_baja, 'HU-10 · Archivar tarjetas terminadas', 'Como usuario, quiero archivar las tarjetas terminadas en lugar de borrarlas, para conservar su historial sin saturar la columna Hecho.', 2, 'low', '2026-11-20 16:14:00');
SET @hu10 := LAST_INSERT_ID();

INSERT INTO card_labels (card_id, label_id) VALUES
 (@hu01, @l_po), (@hu02, @l_po), (@hu03, @l_po), (@hu04, @l_po), (@hu05, @l_po),
 (@hu06, @l_eq), (@hu07, @l_eq), (@hu08, @l_eq), (@hu09, @l_eq), (@hu10, @l_eq),
 (@hu01, @l_s1), (@hu02, @l_s1), (@hu03, @l_s1), (@hu04, @l_s1);

INSERT INTO card_events (card_id, to_column_id, to_kind, moved_at)
SELECT c.id, c.column_id, 'normal', c.created_at FROM cards c WHERE c.id BETWEEN @hu01 AND @hu10;

-- ============ Sprint 1 ============
INSERT INTO boards (name, description, created_at)
VALUES ('Sprint 1', 'Objetivo: que el equipo pueda crear tableros, columnas con límite de WIP y tarjetas, y moverlas sin rebasar ese límite (HU-01 a HU-04).', '2026-11-20 17:00:00');
SET @sb := LAST_INSERT_ID();

INSERT INTO columns_ (board_id, title, position, wip_limit, kind) VALUES (@sb, 'Por hacer', 0, NULL, 'normal');
SET @c_todo := LAST_INSERT_ID();
INSERT INTO columns_ (board_id, title, position, wip_limit, kind) VALUES (@sb, 'En proceso', 1, 3, 'active');
SET @c_doing := LAST_INSERT_ID();
INSERT INTO columns_ (board_id, title, position, wip_limit, kind) VALUES (@sb, 'Hecho', 2, NULL, 'done');
SET @c_done := LAST_INSERT_ID();

INSERT INTO labels (board_id, name, color) VALUES (@sb, 'HU-01', '#2f6fed');
SET @s1 := LAST_INSERT_ID();
INSERT INTO labels (board_id, name, color) VALUES (@sb, 'HU-02', '#e08a1e');
SET @s2 := LAST_INSERT_ID();
INSERT INTO labels (board_id, name, color) VALUES (@sb, 'HU-03', '#8250df');
SET @s3 := LAST_INSERT_ID();
INSERT INTO labels (board_id, name, color) VALUES (@sb, 'HU-04', '#d64545');
SET @s4 := LAST_INSERT_ID();

-- Tareas: (columna final, posición, título, descripción, prioridad, creada)
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_done, 'HU-01 · T1 Tabla de tableros y migración', 'Crear la tabla boards (nombre, descripción, fechas) y su migración.', 0, 'high', '2026-11-20 17:05:00');
SET @t11 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_done, 'HU-01 · T2 API para crear, editar y eliminar tableros', 'Endpoints GET/POST /boards y GET/PATCH/DELETE /boards/:id con validación del nombre.', 1, 'high', '2026-11-20 17:06:00');
SET @t12 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_done, 'HU-01 · T3 Pantalla de tableros', 'Pestañas para listar, crear, editar y eliminar tableros, con confirmación al borrar.', 2, 'high', '2026-11-20 17:07:00');
SET @t13 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_done, 'HU-02 · T1 Columnas con posición y límite de WIP', 'Tabla de columnas con board_id, posición y wip_limit opcional.', 3, 'high', '2026-11-20 17:08:00');
SET @t21 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_done, 'HU-03 · T1 API de tarjetas', 'Crear, editar y eliminar tarjetas con prioridad y fecha de creación asignada por el servidor.', 4, 'high', '2026-11-20 17:09:00');
SET @t31 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_doing, 'HU-02 · T2 Validar el límite de WIP', 'La API rechaza límites menores a 1; vacío significa sin límite.', 0, 'high', '2026-11-20 17:10:00');
SET @t22 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_doing, 'HU-03 · T2 Formulario rápido en cada columna', 'Agregar una tarjeta desde el pie de la columna con solo escribir el título.', 1, 'medium', '2026-11-20 17:11:00');
SET @t32 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_doing, 'HU-04 · T1 Endpoint para mover tarjetas', 'PUT /cards/:id/move con columna y posición destino, reordenando en una transacción.', 2, 'high', '2026-11-20 17:12:00');
SET @t41 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_todo, 'HU-02 · T3 Panel para editar nombre y límite', 'Botón ⚙ en cada columna para cambiar su nombre, límite de WIP y tipo.', 0, 'medium', '2026-11-20 17:13:00');
SET @t23 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_todo, 'HU-03 · T3 Modal de detalle de la tarjeta', 'Ver y editar título, descripción y prioridad; mostrar la fecha de creación.', 1, 'medium', '2026-11-20 17:14:00');
SET @t33 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_todo, 'HU-04 · T2 Regla de WIP en el dominio', 'Responder 409 wip_limit al mover o crear en una columna llena.', 2, 'high', '2026-11-20 17:15:00');
SET @t42 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_todo, 'HU-04 · T3 WIP correcto con peticiones simultáneas', 'Bloquear las columnas en la transacción para que el límite se respete aunque lleguen varios movimientos a la vez.', 3, 'high', '2026-11-20 17:16:00');
SET @t43 := LAST_INSERT_ID();
INSERT INTO cards (column_id, title, description, position, priority, created_at) VALUES
 (@c_todo, 'HU-04 · T4 Arrastrar tarjetas con aviso de columna llena', 'Drag & drop entre columnas; si la columna está llena se revierte y se muestra el aviso.', 4, 'medium', '2026-11-20 17:17:00');
SET @t44 := LAST_INSERT_ID();

INSERT INTO card_labels (card_id, label_id) VALUES
 (@t11, @s1), (@t12, @s1), (@t13, @s1),
 (@t21, @s2), (@t22, @s2), (@t23, @s2),
 (@t31, @s3), (@t32, @s3), (@t33, @s3),
 (@t41, @s4), (@t42, @s4), (@t43, @s4), (@t44, @s4);

-- Historial: todas nacen en Por hacer; las que avanzaron pasan por En proceso (y Hecho).
INSERT INTO card_events (card_id, from_column_id, to_column_id, to_kind, moved_at)
SELECT c.id, NULL, @c_todo, 'normal', c.created_at FROM cards c WHERE c.id BETWEEN @t11 AND @t44;

INSERT INTO card_events (card_id, from_column_id, to_column_id, to_kind, moved_at) VALUES
 (@t11, @c_todo, @c_doing, 'active', '2026-11-20 17:30:00'),
 (@t11, @c_doing, @c_done, 'done', '2026-11-20 19:10:00'),
 (@t12, @c_todo, @c_doing, 'active', '2026-11-20 17:35:00'),
 (@t21, @c_todo, @c_doing, 'active', '2026-11-20 18:00:00'),
 (@t12, @c_doing, @c_done, 'done', '2026-11-20 21:40:00'),
 (@t13, @c_todo, @c_doing, 'active', '2026-11-20 21:45:00'),
 (@t21, @c_doing, @c_done, 'done', '2026-11-20 22:20:00'),
 (@t31, @c_todo, @c_doing, 'active', '2026-11-20 22:25:00'),
 (@t13, @c_doing, @c_done, 'done', '2026-11-21 15:30:00'),
 (@t22, @c_todo, @c_doing, 'active', '2026-11-21 15:35:00'),
 (@t31, @c_doing, @c_done, 'done', '2026-11-21 16:50:00'),
 (@t32, @c_todo, @c_doing, 'active', '2026-11-21 16:55:00'),
 (@t41, @c_todo, @c_doing, 'active', '2026-11-21 17:20:00');
