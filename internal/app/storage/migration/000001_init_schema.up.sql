create table if not exists shortener(
    id serial primary key ,
    short_url varchar(255) unique not null,
    original_url varchar(255));