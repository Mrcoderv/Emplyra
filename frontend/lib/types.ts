export type ApiEnvelope<T> = { success: boolean; message?: string; data: T; errors?: Record<string, string> | string[] }
export type Session = { access_token: string; refresh_token: string; expires_in: number; token_type: string }
export type Scope = 'platform' | 'tenant'
export type CurrentUser = { id: string; email: string; username?: string; first_name: string; last_name: string; role?: string | { name: string }; permissions?: string[] }
export type MeTenant = { id: string; name: string; status: string; plan: string }
export type MeResponse = { user: CurrentUser; permissions: string[]; roles: string[]; scope: Scope; tenant?: MeTenant }
export type DashboardSummary = Record<string, unknown> & {
  total_employees?: number
  active_employees?: number
  employees_on_leave?: number
  present_today?: number
  pending_leave_requests?: number
  open_positions?: number
}
export type ListResponse<T> = { items: T[]; total: number; page: number; page_size: number; total_pages: number }
export type ModuleRecord = Record<string, unknown> & { id?: string; name?: string; title?: string; code?: string; status?: string; created_at?: string }
export type RequestOptions = { page?: number; pageSize?: number; search?: string; query?: Record<string, string | number | boolean | undefined> }
export type CreateEmployee = Record<string, unknown>
export type UpdateEmployee = Record<string, unknown>
export const OPERATIONAL_TENANT_STATUSES = ['ACTIVE', 'TRIAL']

export type Employee = {
  id: string
  employee_code?: string
  first_name: string
  last_name: string
  email?: string
  phone?: string
  date_of_birth?: string
  gender?: string
  address?: string
  emergency_contact?: string
  joining_date?: string
  employment_type?: string
  status?: string
  department_id?: string
  department?: { id: string; name: string } | string | null
  designation_id?: string
  designation?: { id: string; name: string } | string | null
  manager_id?: string
  manager?: { id: string; first_name: string; last_name: string } | string | null
  user_id?: string
  tenant_id?: string
  created_at?: string
  updated_at?: string
}

export type Department = {
  id?: string
  name: string
  code?: string
  description?: string
  manager_id?: string
  status?: string
  tenant_id?: string
  created_at?: string
  updated_at?: string
}

export type Designation = {
  id?: string
  name: string
  description?: string
  department_id?: string
  department?: { id: string; name: string } | string | null
  level?: number
  status?: string
}

export type Holiday = {
  id?: string
  name: string
  date: string
  description?: string
  type?: string
  status?: string
}

export type LeaveType = { id: string; name: string; code?: string; description?: string; is_paid?: boolean }

export type Leave = {
  id?: string
  employee_id?: string
  employee?: Employee | null
  leave_type_id?: string
  leave_type?: LeaveType | null
  start_date?: string
  end_date?: string
  days?: number
  reason?: string
  status?: string
  review_note?: string
  created_at?: string
}

export type LeaveBalance = { id?: string; employee_id?: string; leave_type_id?: string; year?: number; entitlement?: number; used?: number; balance?: number }

export type Document = {
  id?: string
  employee_id?: string
  employee?: Employee | null
  title?: string
  type?: string
  mime_type?: string
  size_bytes?: number
  status?: string
  created_at?: string
}

export type Notification = {
  id?: string
  user_id?: string
  title?: string
  message?: string
  type?: string
  is_read?: boolean
  read_at?: string
  link?: string
  created_at?: string
}

export type SalaryStructure = {
  id?: string
  employee_id?: string
  employee?: Employee | null
  basic_salary?: string
  allowances?: string
  bonus?: string
  overtime_rate?: string
  tax_rate?: string
  tax_amount?: string
  deductions?: string
  effective_from?: string
  effective_until?: string
  status?: string
}

export type Payroll = {
  id?: string
  month?: number
  year?: number
  employee_id?: string
  employee?: Employee | null
  salary_structure_id?: string
  basic_salary?: string
  allowances?: string
  bonus?: string
  overtime?: string
  gross_salary?: string
  tax?: string
  deductions?: string
  net_salary?: string
  status?: string
  payment_ref?: string
  notes?: string
  processed_at?: string
  paid_on?: string
}

export type JobPost = {
  id?: string
  title?: string
  department_id?: string
  department?: { id: string; name: string } | string | null
  description?: string
  requirements?: string
  vacancies?: number
  status?: string
  posted_by?: string
  posted_by_user?: { id: string; first_name: string; last_name: string } | null
  deadline?: string
}

export type Candidate = {
  id?: string
  first_name?: string
  last_name?: string
  email?: string
  phone?: string
  source?: string
  status?: string
  notes?: string
  address?: string
  education?: string
  experience?: string
  skills?: string
}

export type Application = {
  id?: string
  job_post_id?: string
  job_post?: JobPost | null
  candidate_id?: string
  candidate?: Candidate | null
  applied_date?: string
  cover_letter?: string
  status?: string
}

export type Interview = {
  id?: string
  application_id?: string
  application?: Application | null
  interviewer_id?: string
  interviewer?: { id: string; first_name: string; last_name: string } | null
  scheduled_at?: string
  duration_minutes?: number
  type?: string
  status?: string
  feedback?: string
  score?: number | null
}

export type Onboarding = {
  id?: string
  employee_id?: string
  employee?: Employee | null
  candidate_id?: string
  candidate?: Candidate | null
  start_date?: string
  status?: string
  tasks?: string[]
  notes?: string
}

export type Goal = {
  id?: string
  employee_id?: string
  employee?: Employee | null
  title?: string
  description?: string
  target_date?: string
  weight?: number
  status?: string
}

export type KPI = {
  id?: string
  employee_id?: string
  employee?: Employee | null
  name?: string
  description?: string
  target?: string
  actual?: string
  unit?: string
  weight?: number
  period?: string
  score?: number | null
}

export type PerformanceReview = {
  id?: string
  employee_id?: string
  employee?: Employee | null
  reviewer_id?: string
  reviewer?: Employee | null
  period?: string
  self_evaluation?: string
  manager_feedback?: string
  score?: number | null
  status?: string
  due_date?: string
}

export type TrainingProgram = {
  id?: string
  title?: string
  description?: string
  provider?: string
  start_date?: string
  end_date?: string
  location?: string
  max_seats?: number
  status?: string
}

export type TrainingSchedule = {
  id?: string
  program_id?: string
  program?: TrainingProgram | null
  date?: string
  start_time?: string
  end_time?: string
  trainer?: string
  location?: string
  max_seats?: number
}

export type TrainingEnrollment = {
  id?: string
  program_id?: string
  program?: TrainingProgram | null
  employee_id?: string
  employee?: Employee | null
  status?: string
}

export type User = {
  id?: string
  username?: string
  email?: string
  first_name?: string
  last_name?: string
  status?: string
  role_id?: string
  role?: string
  last_login?: string
}

export type Role = {
  id?: string
  name?: string
  description?: string
  is_system?: boolean
  permissions?: { id: string; name: string; description?: string; module?: string }[]
}

export type Team = { id?: string; name?: string; tenant_id?: string }