-- Seed data for PulseWatch

-- Default Admin User (Password: Password123!)
INSERT INTO users (id, email, password_hash, full_name, role)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'admin@pulsewatch.io',
    '$2a$10$eE0m/Z04N77t1V3mF7n5EuK0Q8g3s4u/7Z9y2X1W0v.3u2t1s0r.O',
    'System Admin',
    'superadmin'
) ON CONFLICT (email) DO NOTHING;

-- Default Organization
INSERT INTO organizations (id, name, slug, plan)
VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'Acme Corp',
    'acme-corp',
    'enterprise'
) ON CONFLICT (slug) DO NOTHING;

-- Admin User Membership (Owner)
INSERT INTO members (id, org_id, user_id, role)
VALUES (
    'c0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'owner'
) ON CONFLICT (org_id, user_id) DO NOTHING;

-- Default Project
INSERT INTO projects (id, org_id, name, slug, description, is_public_status_page, status_page_slug)
VALUES (
    'd0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'Production Microservices',
    'prod-services',
    'Core production APIs and infrastructure services monitoring',
    TRUE,
    'acme-prod-status'
) ON CONFLICT (org_id, slug) DO NOTHING;

-- Sample Monitors
INSERT INTO monitors (id, project_id, name, type, url, method, interval_seconds, expected_status_code, status)
VALUES
(
    'e0000000-0000-0000-0000-000000000001',
    'd0000000-0000-0000-0000-000000000001',
    'Google Search Gateway',
    'http',
    'https://www.google.com',
    'GET',
    30,
    200,
    'up'
),
(
    'e0000000-0000-0000-0000-000000000002',
    'd0000000-0000-0000-0000-000000000001',
    'Cloudflare DNS Resolver',
    'dns',
    '1.1.1.1',
    'GET',
    60,
    200,
    'up'
),
(
    'e0000000-0000-0000-0000-000000000003',
    'd0000000-0000-0000-0000-000000000001',
    'GitHub SSL Expiry Check',
    'ssl',
    'https://github.com',
    'GET',
    300,
    200,
    'up'
)
ON CONFLICT (id) DO NOTHING;

-- Sample Check Results
INSERT INTO check_results (monitor_id, status, status_code, latency_ms, dns_time_ms, ssl_days_remaining)
VALUES
('e0000000-0000-0000-0000-000000000001', 'up', 200, 45, 12, NULL),
('e0000000-0000-0000-0000-000000000001', 'up', 200, 52, 10, NULL),
('e0000000-0000-0000-0000-000000000002', 'up', 200, 15, 8, NULL),
('e0000000-0000-0000-0000-000000000003', 'up', 200, 120, 20, 180);

-- Sample Notification Target
INSERT INTO notifications (id, project_id, type, target, is_enabled)
VALUES (
    'f0000000-0000-0000-0000-000000000001',
    'd0000000-0000-0000-0000-000000000001',
    'webhook',
    'https://webhook.site/pulsewatch-alerts',
    TRUE
) ON CONFLICT (id) DO NOTHING;
