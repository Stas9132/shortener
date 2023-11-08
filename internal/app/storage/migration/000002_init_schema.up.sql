create table if not exists shortener(
    id serial primary key,
    user_id varchar(255),
    short_url varchar(255) unique not null,
    original_url varchar(255),
    is_deleted boolean);