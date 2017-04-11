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