CREATE TABLE IF NOT EXISTS signals (
	id INT PRIMARY KEY NOT NULL SERIAL, 
	name TEXT NOT NULL UNIQUE CHECK (name <> ''), 
	description TEXT, 
	num_subscribers INT CONSTRAINT non_negative_num_subscribers CHECK (num_subscribers >= 0), 
	num_trades INT CONSTRAINT non_negative_num_trades CHECK (num_trades >= 0), 
	price REAL CONSTRAINT positive_price CHECK (price > 0), 
	growth REAL
);

CREATE TABLE IF NOT EXISTS users (id SERIAL UNIQUE, email TEXT NOT NULL UNIQUE CHECK (email <> ''), password TEXT NOT NULL CHECK (password <> ''));

CREATE TABLE IF NOT EXISTS closed_orders (id SERIAL UNIQUE, email TEXT NOT NULL UNIQUE CHECK (email <> ''), password TEXT NOT NULL CHECK (password <> ''));

CREATE TABLE IF NOT EXISTS holdings (id SERIAL UNIQUE, signal_id INT REFERENCES signals(id), code TEXT NOT NULL CHECK (code <> ''), name TEXT NOT NULL CHECK (code <> ''), num_shares INT CONSTRAINT non_negative_num_shares CHECK (num_shares >= 0), price DECIMAL(7,2) CONSTRAINT positive_price CHECK (price > 0));

CREATE TABLE IF NOT EXISTS orders (id SERIAL UNIQUE, signal_id INT REFERENCES signals(id),  order_time timestamp, type TEXT NOT NULL CHECK (code <> ''), code TEXT NOT NULL CHECK (code <> ''), name TEXT NOT NULL CHECK (code <> ''), num_shares INT CONSTRAINT non_negative_num_shares CHECK (num_shares >= 0), price DECIMAL(7,2) CONSTRAINT positive_price CHECK (price > 0), profit DECIMAL(7,2));
