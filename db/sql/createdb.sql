create database diego;

\connect diego;

create table domains (
  domain varchar(256),
  expireTime timestamp
);

create table actuals (
  processGuid varchar(64),
  idx int,
  domain varchar(256),
  cellId varchar(256),
  isEvacuating boolean,
  data bytea
);
