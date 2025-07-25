-- Create initial admin user
-- Password: admin123 (hashed with bcrypt)
INSERT INTO users (username, password_hash, full_name, email, role, is_active)
VALUES (
    'admin',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- admin123
    'System Administrator',
    'admin@digitalsignature.com',
    'admin',
    true
) ON CONFLICT (username) DO NOTHING;

-- Create initial regular user for testing
-- Password: user123 (hashed with bcrypt)
INSERT INTO users (username, password_hash, full_name, email, role, is_active)
VALUES (
    'testuser',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- user123
    'Test User',
    'test@digitalsignature.com',
    'user',
    true
) ON CONFLICT (username) DO NOTHING;