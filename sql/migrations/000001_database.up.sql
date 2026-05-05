CREATE TABLE ride_status (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE vehicles (
    id UUID PRIMARY KEY,
    license_plate VARCHAR(20) UNIQUE NOT NULL,
    model VARCHAR(100) NOT NULL,
    brand VARCHAR(100) NOT NULL,
    color VARCHAR(50) NOT NULL
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(200) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    photo_url VARCHAR(200) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE drivers (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL,
    cnh_number VARCHAR(20) UNIQUE NOT NULL,
    vehicle_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
);

CREATE TABLE rides (
    id UUID PRIMARY KEY DEFAULT 'c2e85aee-be95-4f38-ad0b-dd096b78bac0',
    passenger_id UUID NOT NULL,
    driver_id UUID,

    origin_address VARCHAR(255) NOT NULL,
    origin_lat DOUBLE PRECISION NOT NULL,
    origin_lng DOUBLE PRECISION NOT NULL,

    destination_address VARCHAR(255) NOT NULL,
    destination_lat DOUBLE PRECISION NOT NULL,
    destination_lng DOUBLE PRECISION NOT NULL,

    status_id UUID NOT NULL,

    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    monetary_value DECIMAL(10,2),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (driver_id) REFERENCES drivers(id),
    FOREIGN KEY (status_id) REFERENCES ride_status(id)
);