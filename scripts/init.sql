-- ============================================================
-- Script de inicialización de la base de datos
-- Se ejecuta automáticamente al crear el contenedor MySQL
-- ============================================================

CREATE DATABASE IF NOT EXISTS apidb;
USE apidb;

-- ------------------------------------------------------------
-- Tabla: users
-- Almacena hosts (profesores) y participantes (alumnos)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id         INT AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(100)        NOT NULL,
    email      VARCHAR(150)        NOT NULL UNIQUE,
    password   VARCHAR(255)        NOT NULL,
    role       ENUM('host','participant') NOT NULL DEFAULT 'participant',
    created_at TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ------------------------------------------------------------
-- Tabla: rooms
-- Salas creadas por los hosts
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS rooms (
    id         INT AUTO_INCREMENT PRIMARY KEY,
    code       VARCHAR(10)         NOT NULL UNIQUE,
    host_id    INT                 NOT NULL,
    status     ENUM('waiting','active','finished') NOT NULL DEFAULT 'waiting',
    created_at TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_rooms_host FOREIGN KEY (host_id) REFERENCES users(id) ON DELETE CASCADE
);

-- ------------------------------------------------------------
-- Tabla: participants
-- Relación de qué usuarios están en qué sala
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS participants (
    id         INT AUTO_INCREMENT PRIMARY KEY,
    room_id    INT NOT NULL,
    user_id    INT NOT NULL,
    joined_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_part_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    CONSTRAINT fk_part_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_room_user (room_id, user_id)
);

-- ------------------------------------------------------------
-- Tabla: scores
-- Puntos de cada participante dentro de una sala
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS scores (
    id         INT AUTO_INCREMENT PRIMARY KEY,
    room_id    INT NOT NULL,
    user_id    INT NOT NULL,
    points     INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_score_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    CONSTRAINT fk_score_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_score_room_user (room_id, user_id)
);

-- ------------------------------------------------------------
-- Tabla: questions
-- Preguntas lanzadas por el host durante una sesión activa
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS questions (
    id             INT AUTO_INCREMENT PRIMARY KEY,
    room_id        INT NOT NULL,
    text           TEXT NOT NULL,
    correct_answer VARCHAR(500) NOT NULL,
    points         INT NOT NULL DEFAULT 10,
    status         ENUM('open','closed') NOT NULL DEFAULT 'open',
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_question_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
);

-- ------------------------------------------------------------
-- Tabla: answers
-- Respuestas enviadas por los participantes a las preguntas
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS answers (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    question_id INT NOT NULL,
    user_id     INT NOT NULL,
    text        VARCHAR(500) NOT NULL,
    is_correct  BOOLEAN NOT NULL DEFAULT FALSE,
    answered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_answer_question FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE,
    CONSTRAINT fk_answer_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_answer_question_user (question_id, user_id)  -- un participante solo responde una vez
);
