drop database diego;

create database diego;

use diego;

create table domains (
  domain varchar(256),
  expireTime timestamp
);

create table actuals (
  processGuid varchar(80),
  idx int,
  domain varchar(256),
  cellId varchar(256),
  isEvacuating boolean,
  modifiedIndex int,
  data blob
);

create index on actuals (cellId);

create table desired (
  processGuid varchar(80),
  domain varchar(256),
  modifiedIndex int,
  scheduleInfo blob,
  runInfo blob
);
