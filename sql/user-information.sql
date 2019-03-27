CREATE DATABASE IF NOT EXISTS authorization_info;

USE authorization_info;

DROP TABLE IF EXISTS USERS_GROUPS;
DROP TABLE IF EXISTS USERS;
DROP TABLE IF EXISTS GROUPS;

CREATE TABLE IF NOT EXISTS USERS(
  id VARCHAR(255),
  name VARCHAR(255) NOT NULL,
  house VARCHAR(255) NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS GROUPS(
  id VARCHAR(255),
  name VARCHAR(255) NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS USERS_GROUPS(
  id_user VARCHAR(255),
  id_group VARCHAR(255),
  INDEX id_user_ind (id_user),
  INDEX id_group_ind (id_group),
  PRIMARY KEY (id_user, id_group),
  FOREIGN KEY (id_user)
    REFERENCES USERS(id)
    ON DELETE CASCADE,
  FOREIGN KEY (id_group)
    REFERENCES GROUPS(id)
    ON DELETE CASCADE
);

INSERT INTO USERS (id, name, house) VALUES ('thawat', 'Thufir Hawat', 'Atreides');
INSERT INTO USERS (id, name, house) VALUES ('patreides', 'Paul Atreides', 'Atreides');
INSERT INTO USERS (id, name, house) VALUES ('vharkonen', 'Vladimir Harkonen', 'Harkonen');
INSERT INTO USERS (id, name, house) VALUES ('pdevries', 'Piter De Vries', 'Harkonen');
INSERT INTO GROUPS(id, name) VALUES ('moa', 'Master of Assassins');
INSERT INTO GROUPS(id, name) VALUES ('mentat', 'Mentat');
INSERT INTO GROUPS(id, name) VALUES ('duke', 'Duke');
INSERT INTO GROUPS(id, name) VALUES ('baron', 'Baron');
INSERT INTO GROUPS(id, name) VALUES ('kh', 'Kwisatz Haderach');
INSERT INTO USERS_GROUPS(id_user, id_group) VALUES ('thawat', 'moa');
INSERT INTO USERS_GROUPS(id_user, id_group) VALUES ('thawat', 'mentat');
INSERT INTO USERS_GROUPS(id_user, id_group) VALUES ('patreides', 'duke');
INSERT INTO USERS_GROUPS(id_user, id_group) VALUES ('patreides', 'kh');
INSERT INTO USERS_GROUPS(id_user, id_group) VALUES ('vharkonen', 'baron');
INSERT INTO USERS_GROUPS(id_user, id_group) VALUES ('pdevries', 'mentat');

select USERS.id Username, USERS.name Name, USERS.house House, GROUPS.name 'Group' from USERS JOIN USERS_GROUPS JOIN GROUPS on USERS.id = USERS_GROUPS.id_user and GROUPS.id = USERS_GROUPS.id_group ORDER BY USERS.house, GROUPS.name;

