
CREATE TABLE IF NOT EXISTS public.junctions
(
    junction_id bigint NOT NULL DEFAULT nextval('junctions_junction_id_seq'::regclass),
    name text COLLATE pg_catalog."default" NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    status junctions_status_enum NOT NULL DEFAULT 'active'::junctions_status_enum,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    city text COLLATE pg_catalog."default",
    CONSTRAINT junctions_pkey PRIMARY KEY (junction_id)
);


CREATE TABLE IF NOT EXISTS public.cameras
(
    camera_id bigint NOT NULL DEFAULT nextval('cameras_camera_id_seq'::regclass),
    junction_id bigint NOT NULL,
    rtsp_url text COLLATE pg_catalog."default" NOT NULL,
    angle double precision NOT NULL,
    resolution character varying(18) COLLATE pg_catalog."default" NOT NULL,
    status cameras_status_enum NOT NULL DEFAULT 'active'::cameras_status_enum,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    policy_id bigint NOT NULL DEFAULT 10000000,
    CONSTRAINT cameras_pkey PRIMARY KEY (camera_id),
    CONSTRAINT junctions_cameras_junction_id_fkey FOREIGN KEY (junction_id)
        REFERENCES public.junctions (junction_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID,
    CONSTRAINT policies_cameras_policy_id_fkey FOREIGN KEY (policy_id)
        REFERENCES public.policies (policy_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID
);

CREATE TABLE IF NOT EXISTS public.policies
(
    policy_id bigint NOT NULL DEFAULT nextval('policies_policy_id_seq'::regclass),
    green_min integer NOT NULL DEFAULT 30,
    green_max integer NOT NULL DEFAULT 90,
    red_min integer NOT NULL DEFAULT 30,
    red_max integer NOT NULL DEFAULT 60,
    cycle_max integer NOT NULL DEFAULT 180,
    CONSTRAINT policies_pkey PRIMARY KEY (policy_id)
);
