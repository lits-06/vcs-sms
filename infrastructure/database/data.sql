--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5 (Debian 17.5-1.pgdg120+1)
-- Dumped by pg_dump version 17.5 (Debian 17.5-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: servers; Type: TABLE; Schema: public; Owner: dev_user
--

CREATE TABLE public.servers (
    id text NOT NULL,
    name text,
    status text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    ipv4 text
);


ALTER TABLE public.servers OWNER TO dev_user;

--
-- Data for Name: servers; Type: TABLE DATA; Schema: public; Owner: dev_user
--

COPY public.servers (id, name, status, created_at, updated_at, ipv4) FROM stdin;
lb-4014	Cache Server 87 (Development - EU-Central)	OFF	2025-08-01 08:35:03.242903+00	2025-08-01 08:35:03.242903+00	172.16.26.251
web-4634	Cache Server 33 (Testing - Canada)	OFF	2025-08-01 08:35:03.281754+00	2025-08-01 08:35:03.281754+00	192.168.224.177
srv-2621	Mail Server 63 (Development - Asia-Pacific)	OFF	2025-08-01 08:35:03.287083+00	2025-08-01 08:35:03.287083+00	192.168.197.94
web-6585	Cache Server 99 (Demo - US-East)	OFF	2025-08-01 08:35:03.48359+00	2025-08-01 08:35:03.483591+00	172.16.27.96
cache-4650	Mail Server 49 (Testing - Canada)	OFF	2025-08-01 08:35:03.488834+00	2025-08-01 08:35:03.488834+00	10.117.96.233
file-1980	Load Balancer Server 90 (Production - US-West)	OFF	2025-08-01 08:35:03.498022+00	2025-08-01 08:35:03.498022+00	172.16.23.173
web-2985	Web Server 60 (Testing - Canada)	OFF	2025-08-01 08:35:03.254784+00	2025-08-01 08:35:06.493138+00	172.16.22.220
db-2435	Web Server 80 (Demo - Australia)	OFF	2025-08-01 08:35:03.257758+00	2025-08-01 08:35:06.506126+00	192.168.235.39
lb-4728	Web Server 27 (Production - US-East)	OFF	2025-08-01 08:35:03.294233+00	2025-08-01 08:35:06.510651+00	192.168.153.189
cache-9233	Load Balancer Server 45 (Production - Australia)	OFF	2025-08-01 08:35:03.480745+00	2025-08-01 08:35:06.514278+00	172.16.23.99
\.


--
-- Name: servers servers_pkey; Type: CONSTRAINT; Schema: public; Owner: dev_user
--

ALTER TABLE ONLY public.servers
    ADD CONSTRAINT servers_pkey PRIMARY KEY (id);


--
-- Name: idx_servers_name; Type: INDEX; Schema: public; Owner: dev_user
--

CREATE UNIQUE INDEX idx_servers_name ON public.servers USING btree (name);


--
-- PostgreSQL database dump complete
--

