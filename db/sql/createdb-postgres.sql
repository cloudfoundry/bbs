create database diego;

\connect diego;

create table domains (
  domain varchar(256),
  expireTime timestamp
);

grant all on domains to public;

create table actuals (
  processGuid varchar(64),
  idx int,
  domain varchar(256),
  cellId varchar(256),
  isEvacuating boolean,
  modifiedIndex int,
  data bytea
);

grant all on domains to public;

create index on actuals (cellId);

create table desired (
  processGuid varchar(64),
  domain varchar(256),
  scheduleInfo bytea,
  modifiedIndex int,
  runInfo bytea
);

grant all on actuals to public;
