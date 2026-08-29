-- HRMS initial schema. The application also runs GORM AutoMigrate on startup
-- so this file is optional and idempotent.

-- Core identity tables
CREATE TABLE IF NOT EXISTS permissions (
    id uuid PRIMARY KEY,
    name varchar(100) NOT NULL UNIQUE,
    module varchar(50) NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS roles (
    id uuid PRIMARY KEY,
    name varchar(100) NOT NULL UNIQUE,
    description text NOT NULL DEFAULT '',
    is_system boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    username varchar(100) NOT NULL UNIQUE,
    email varchar(255) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,
    first_name varchar(100) NOT NULL,
    last_name varchar(100) NOT NULL DEFAULT '',
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    role_id uuid NOT NULL REFERENCES roles(id),
    last_login_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash varchar(64) NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    replaced_by_hash varchar(64),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id uuid PRIMARY KEY,
    user_id uuid REFERENCES users(id),
    action varchar(40) NOT NULL,
    resource varchar(100) NOT NULL DEFAULT '',
    resource_id varchar(64) NOT NULL DEFAULT '',
    ip_address varchar(64) NOT NULL DEFAULT '',
    user_agent varchar(255) NOT NULL DEFAULT '',
    metadata jsonb,
    created_at timestamptz NOT NULL
);

-- Organization
CREATE TABLE IF NOT EXISTS departments (
    id uuid PRIMARY KEY,
    name varchar(150) NOT NULL,
    code varchar(50) NOT NULL UNIQUE,
    description varchar(500) NOT NULL DEFAULT '',
    manager_id uuid,
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS designations (
    id uuid PRIMARY KEY,
    name varchar(150) NOT NULL,
    description varchar(500) NOT NULL DEFAULT '',
    department_id uuid REFERENCES departments(id),
    level integer NOT NULL DEFAULT 1,
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS employees (
    id uuid PRIMARY KEY,
    employee_code varchar(50) NOT NULL UNIQUE,
    first_name varchar(100) NOT NULL,
    last_name varchar(100) NOT NULL DEFAULT '',
    email varchar(255) NOT NULL UNIQUE,
    phone varchar(30) NOT NULL DEFAULT '',
    date_of_birth date,
    gender varchar(20),
    address varchar(500) NOT NULL DEFAULT '',
    emergency_contact varchar(255) NOT NULL DEFAULT '',
    joining_date date,
    employment_type varchar(20) NOT NULL DEFAULT 'FULL_TIME',
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    department_id uuid REFERENCES departments(id),
    designation_id uuid REFERENCES designations(id),
    manager_id uuid REFERENCES employees(id),
    user_id uuid REFERENCES users(id),
    deleted_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- Time off
CREATE TABLE IF NOT EXISTS leave_types (
    id uuid PRIMARY KEY,
    name varchar(100) NOT NULL,
    code varchar(50) NOT NULL UNIQUE,
    description varchar(500) NOT NULL DEFAULT '',
    is_paid boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS leave_balances (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL REFERENCES employees(id),
    leave_type_id uuid NOT NULL REFERENCES leave_types(id),
    year integer NOT NULL,
    entitlement integer NOT NULL DEFAULT 0,
    used integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS leaves (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL REFERENCES employees(id),
    leave_type_id uuid NOT NULL REFERENCES leave_types(id),
    start_date date NOT NULL,
    end_date date NOT NULL,
    days integer NOT NULL DEFAULT 0,
    reason varchar(1000) NOT NULL DEFAULT '',
    status varchar(20) NOT NULL DEFAULT 'PENDING',
    reviewer_id uuid REFERENCES employees(id),
    reviewed_at timestamptz,
    review_note varchar(1000) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS holidays (
    id uuid PRIMARY KEY,
    name varchar(150) NOT NULL,
    date date NOT NULL,
    description varchar(500) NOT NULL DEFAULT '',
    type varchar(50) NOT NULL DEFAULT '',
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS attendances (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL REFERENCES employees(id),
    date date NOT NULL,
    check_in timestamptz,
    check_out timestamptz,
    working_hours double precision NOT NULL DEFAULT 0,
    status varchar(20) NOT NULL DEFAULT 'PRESENT',
    late_minutes integer NOT NULL DEFAULT 0,
    overtime double precision NOT NULL DEFAULT 0,
    remarks varchar(500) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT uq_attendance_employee_date UNIQUE (employee_id, date)
);

-- Payroll
CREATE TABLE IF NOT EXISTS salary_structures (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL REFERENCES employees(id),
    basic_salary numeric(14,2) NOT NULL,
    allowances numeric(14,2) NOT NULL DEFAULT 0,
    bonus numeric(14,2) NOT NULL DEFAULT 0,
    overtime_rate numeric(14,2) NOT NULL DEFAULT 0,
    tax_rate numeric(5,2) NOT NULL DEFAULT 0,
    tax_amount numeric(14,2) NOT NULL DEFAULT 0,
    deductions numeric(14,2) NOT NULL DEFAULT 0,
    effective_from date NOT NULL,
    effective_until date,
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS payrolls (
    id uuid PRIMARY KEY,
    month integer NOT NULL,
    year integer NOT NULL,
    employee_id uuid NOT NULL REFERENCES employees(id),
    salary_structure_id uuid REFERENCES salary_structures(id),
    basic_salary numeric(14,2) NOT NULL,
    allowances numeric(14,2) NOT NULL DEFAULT 0,
    bonus numeric(14,2) NOT NULL DEFAULT 0,
    overtime numeric(14,2) NOT NULL DEFAULT 0,
    gross_salary numeric(14,2) NOT NULL DEFAULT 0,
    tax numeric(14,2) NOT NULL DEFAULT 0,
    deductions numeric(14,2) NOT NULL DEFAULT 0,
    net_salary numeric(14,2) NOT NULL DEFAULT 0,
    status varchar(20) NOT NULL DEFAULT 'DRAFT',
    processed_by uuid REFERENCES users(id),
    processed_at timestamptz,
    paid_on timestamptz,
    cancelled_by uuid REFERENCES users(id),
    cancelled_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT uq_payroll_employee_month UNIQUE (employee_id, month, year)
);

-- Recruitment
CREATE TABLE IF NOT EXISTS job_posts (
    id uuid PRIMARY KEY,
    title varchar(200) NOT NULL,
    department_id uuid REFERENCES departments(id),
    description text NOT NULL DEFAULT '',
    requirements text NOT NULL DEFAULT '',
    vacancies integer NOT NULL DEFAULT 1,
    status varchar(20) NOT NULL DEFAULT 'OPEN',
    posted_by uuid REFERENCES users(id),
    deadline date,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS candidates (
    id uuid PRIMARY KEY,
    first_name varchar(100) NOT NULL,
    last_name varchar(100) NOT NULL DEFAULT '',
    email varchar(255) NOT NULL,
    phone varchar(30) NOT NULL DEFAULT '',
    resume_path varchar(500) NOT NULL DEFAULT '',
    source varchar(100) NOT NULL DEFAULT '',
    status varchar(20) NOT NULL DEFAULT 'NEW',
    notes varchar(2000) NOT NULL DEFAULT '',
    deleted_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS applications (
    id uuid PRIMARY KEY,
    job_post_id uuid NOT NULL REFERENCES job_posts(id),
    candidate_id uuid NOT NULL REFERENCES candidates(id),
    applied_date date NOT NULL,
    cover_letter text NOT NULL DEFAULT '',
    status varchar(20) NOT NULL DEFAULT 'APPLIED',
    reviewer_id uuid REFERENCES users(id),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS interviews (
    id uuid PRIMARY KEY,
    application_id uuid NOT NULL REFERENCES applications(id),
    interviewer_id uuid REFERENCES users(id),
    scheduled_at timestamptz NOT NULL,
    duration_minutes integer NOT NULL DEFAULT 0,
    type varchar(20) NOT NULL DEFAULT '',
    status varchar(20) NOT NULL DEFAULT 'SCHEDULED',
    feedback text NOT NULL DEFAULT '',
    score double precision,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS onboardings (
    id uuid PRIMARY KEY,
    employee_id uuid REFERENCES employees(id),
    candidate_id uuid REFERENCES candidates(id),
    start_date date,
    status varchar(20) NOT NULL DEFAULT 'PENDING',
    notes varchar(2000) NOT NULL DEFAULT '',
    tasks jsonb,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- Performance & training
CREATE TABLE IF NOT EXISTS goals (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL REFERENCES employees(id),
    title varchar(200) NOT NULL,
    description text NOT NULL DEFAULT '',
    target_date date,
    weight double precision NOT NULL DEFAULT 0,
    status varchar(20) NOT NULL DEFAULT 'NOT_STARTED',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS kpis (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL REFERENCES employees(id),
    name varchar(200) NOT NULL,
    description text NOT NULL DEFAULT '',
    target numeric(14,2) NOT NULL DEFAULT 0,
    actual numeric(14,2) NOT NULL DEFAULT 0,
    unit varchar(30) NOT NULL DEFAULT '',
    weight double precision NOT NULL DEFAULT 0,
    period varchar(20) NOT NULL DEFAULT '',
    score double precision,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS performance_reviews (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL REFERENCES employees(id),
    reviewer_id uuid REFERENCES employees(id),
    period varchar(30) NOT NULL,
    due_date date,
    self_evaluation text NOT NULL DEFAULT '',
    manager_feedback text NOT NULL DEFAULT '',
    score double precision,
    status varchar(30) NOT NULL DEFAULT 'PENDING',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS training_programs (
    id uuid PRIMARY KEY,
    title varchar(200) NOT NULL,
    description text NOT NULL DEFAULT '',
    provider varchar(200) NOT NULL DEFAULT '',
    start_date date,
    end_date date,
    location varchar(200) NOT NULL DEFAULT '',
    max_seats integer NOT NULL DEFAULT 0,
    status varchar(20) NOT NULL DEFAULT 'SCHEDULED',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS training_schedules (
    id uuid PRIMARY KEY,
    program_id uuid NOT NULL REFERENCES training_programs(id),
    date date NOT NULL,
    start_time varchar(10) NOT NULL DEFAULT '',
    end_time varchar(10) NOT NULL DEFAULT '',
    trainer varchar(200) NOT NULL DEFAULT '',
    location varchar(200) NOT NULL DEFAULT '',
    max_seats integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS training_enrollments (
    id uuid PRIMARY KEY,
    program_id uuid NOT NULL REFERENCES training_programs(id),
    employee_id uuid NOT NULL REFERENCES employees(id),
    status varchar(20) NOT NULL DEFAULT 'ENROLLED',
    enrolled_at timestamptz NOT NULL,
    completed_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- Documents & notifications
CREATE TABLE IF NOT EXISTS documents (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL REFERENCES employees(id),
    title varchar(200) NOT NULL,
    type varchar(30) NOT NULL DEFAULT 'OTHER',
    file_path varchar(500) NOT NULL,
    mime_type varchar(100) NOT NULL DEFAULT '',
    size_bytes bigint NOT NULL DEFAULT 0,
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    uploaded_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS notifications (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id),
    title varchar(200) NOT NULL,
    message varchar(2000) NOT NULL DEFAULT '',
    type varchar(40) NOT NULL DEFAULT 'SYSTEM',
    is_read boolean NOT NULL DEFAULT false,
    read_at timestamptz,
    link varchar(255) NOT NULL DEFAULT '',
    metadata jsonb,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- Common indexes
CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_employees_designation ON employees(designation_id);
CREATE INDEX IF NOT EXISTS idx_employees_manager ON employees(manager_id);
CREATE INDEX IF NOT EXISTS idx_employees_user ON employees(user_id);
CREATE INDEX IF NOT EXISTS idx_leaves_employee ON leaves(employee_id);
CREATE INDEX IF NOT EXISTS idx_leaves_status ON leaves(status);
CREATE INDEX IF NOT EXISTS idx_attendances_date ON attendances(date);
CREATE INDEX IF NOT EXISTS idx_payrolls_employee_month ON payrolls(employee_id, year, month);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read);