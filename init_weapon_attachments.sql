CREATE TABLE IF NOT EXISTS weapon_attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    weapon_name TEXT NOT NULL,
    attachment_name TEXT NOT NULL
);
DELETE FROM weapon_attachments;

INSERT INTO weapon_attachments (weapon_name, attachment_name) VALUES 
('Virtus', 'Tactical Flashlight'), ('Virtus', 'Tactical Suppressor'), ('Virtus', 'Extended Rifle Mag'), ('Virtus', 'Medium Scope'),
('Vektor KVR', 'Tactical Suppressor'), ('Vektor KVR', 'Grip'), ('Vektor KVR', 'Medium Scope'),
('X17', 'Tactical Flashlight'),
('P50', 'Tactical Suppressor'), ('P50', 'Extended Pistol Mag'),
('Machine Pistol', 'SMG Drum Mag'), ('Machine Pistol', 'Suppressor'),
('Mini SMG', 'Extended SMG Mag'),
('Micro SMG', 'Tactical Suppressor'), ('Micro SMG', 'Tactical Flashlight'), ('Micro SMG', 'Extended SMG Mag'), ('Micro SMG', 'Macro Scope'),
('SMG', 'SMG Drum Mag'), ('SMG', 'Suppressor'), ('SMG', 'Tactical Flashlight'), ('SMG', 'Macro Scope'),
('Shotgun', 'Tactical Suppressor'), ('Shotgun', 'Tactical Flashlight'),
('Carbine', 'Extended Rifle Mag'), ('Carbine', 'Rifle Drum Mag'), ('Carbine', 'Medium Scope'), ('Carbine', 'Grip'), ('Carbine', 'Tactical Suppressor'),
('AK', 'Tactical Flashlight'), ('AK', 'Grip'), ('AK', 'Rifle Drum Mag'), ('AK', 'Extended Rifle Mag'), ('AK', 'Macro Scope'), ('AK', 'Tactical Suppressor'),
('Ceramic', 'Extended Pistol Mag'), ('Ceramic', 'Suppressor');
