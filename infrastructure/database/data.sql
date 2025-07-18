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
    "updated_at,autoUpdateTime" timestamp with time zone,
    ipv4 text
);


ALTER TABLE public.servers OWNER TO dev_user;

--
-- Data for Name: servers; Type: TABLE DATA; Schema: public; Owner: dev_user
--

COPY public.servers (id, name, status, created_at, "updated_at,autoUpdateTime", ipv4) FROM stdin;
lb-1600	Monitor Server 35 (Development - Canada)	OFF	2025-07-18 10:10:28.030155+00	2025-07-18 10:10:28.030155+00	192.168.64.23
cache-5333	Load Balancer Server 83 (Testing - Australia)	ON	2025-07-18 10:10:28.037855+00	2025-07-18 10:10:28.037855+00	192.168.113.222
api-2389	Database Server 08 (Production - US-East)	ON	2025-07-18 10:10:28.037042+00	2025-07-18 10:10:28.037042+00	172.16.16.74
web-9725	DNS Server 39 (Development - EU-Central)	ON	2025-07-18 10:10:28.037549+00	2025-07-18 10:10:28.037549+00	172.16.30.126
mail-1830	Load Balancer Server 84 (Production - Canada)	OFF	2025-07-18 10:10:28.150229+00	2025-07-18 10:10:28.150229+00	192.168.7.6
lb-4021	Mail Server 16 (Production - US-East)	OFF	2025-07-18 10:10:28.151561+00	2025-07-18 10:10:28.151561+00	172.16.16.74
api-0604	API Server 18 (Production - Australia)	OFF	2025-07-18 10:10:28.15536+00	2025-07-18 10:10:28.15536+00	192.168.11.74
mail-9294	Load Balancer Server 53 (Demo - US-East)	OFF	2025-07-18 10:10:28.167042+00	2025-07-18 10:10:28.167043+00	10.62.61.114
file-4210	Web Server 74 (Development - Asia-Pacific)	ON	2025-07-18 10:10:28.165053+00	2025-07-18 10:10:28.165053+00	192.168.209.93
file-9417	API Server 26 (Testing - Asia-Pacific)	ON	2025-07-18 10:10:28.16607+00	2025-07-18 10:10:28.166071+00	192.168.69.170
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

