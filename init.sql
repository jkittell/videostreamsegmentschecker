CREATE TABLE segment_checks (
                       id uuid,
                       url varchar,
                       ok_segments int,
                       total_segments int,
                       ok_percent float,
                       created_at timestamp without time zone
);