'use client'

import Link from 'next/link'
import CrudListPage from '@/components/crud-list-page'
import { api, type Department, type Designation, type Employee } from '@/lib/api'

export default function EmployeesPage() {
  return (
    <CrudListPage<Employee>
      title="Employees"
      eyebrow="Directory"
      description="Find people, roles, and reporting context across your organization."
      actionLabel="Add employee"
      empty="No employees have been added yet."
      fields={[
        { name: 'employee_code', label: 'Employee code', required: true, placeholder: 'e.g. EMP-001' },
        { name: 'first_name', label: 'First name', required: true },
        { name: 'last_name', label: 'Last name', required: true },
        { name: 'email', label: 'Work email', type: 'email', required: true },
        { name: 'phone', label: 'Phone', placeholder: '+977…' },
        { name: 'department_id', label: 'Department', type: 'select', options: [] },
        { name: 'designation_id', label: 'Designation', type: 'select', options: [] },
        { name: 'employment_type', label: 'Employment type', type: 'select', options: ['FULL_TIME', 'PART_TIME', 'CONTRACT', 'INTERN', 'PROBATION'].map((v) => ({ label: v.replaceAll('_', ' '), value: v })) },
        { name: 'joining_date', label: 'Joining date', type: 'date' },
        { name: 'status', label: 'Status', type: 'select', options: ['ACTIVE', 'INACTIVE', 'ON_LEAVE', 'TERMINATED', 'ONBOARDING'].map((v) => ({ label: v.replaceAll('_', ' '), value: v })) },
      ]}
      columns={['department', 'designation', 'employment_type', 'status']}
      load={async (query) => {
        const [employees, deps, desigs] = await Promise.all([
          api.employees(1, 100, query),
          api.departments({ page: 1, pageSize: 100 }),
          api.designations({ page: 1, pageSize: 100 }),
        ])
        const deptMap = new Map((deps.items ?? []).map((d: Department) => [d.id, d.name]))
        const desigMap = new Map((desigs.items ?? []).map((d: Designation) => [d.id, d.name]))
        return (employees.items ?? []).map((e) => ({
          ...e,
          department: e.department_id ? deptMap.get(e.department_id) ?? (e.department && typeof e.department === 'object' ? e.department.name : '—') : '—',
          designation: e.designation_id ? desigMap.get(e.designation_id) ?? (e.designation && typeof e.designation === 'object' ? e.designation.name : '—') : '—',
        }))
      }}
      create={(body) => api.employee.create(body)}
      update={(id, body) => api.employee.update(id, body)}
      remove={(id) => api.employee.remove(id)}
      fieldOptions={{
        department_id: async () => (await api.departments({ page: 1, pageSize: 100 })).items?.map((d) => ({ label: d.name, value: d.id ?? '' })) ?? [],
        designation_id: async () => (await api.designations({ page: 1, pageSize: 100 })).items?.map((d) => ({ label: d.name, value: d.id ?? '' })) ?? [],
      }}
      render={(e) => (
        <>
          <td className="px-5 py-4"><Link className="font-medium hover:text-primary" href={`/employees/${e.id}`}>{e.first_name} {e.last_name}</Link><p className="text-xs text-muted-foreground">{e.email || e.employee_code || 'Employee record'}</p></td>
          <td className="px-5 py-4 text-muted-foreground">{displayName(e.department)}</td>
          <td className="px-5 py-4 text-muted-foreground">{displayName(e.designation)}</td>
          <td className="px-5 py-4 text-muted-foreground">{e.employment_type || '—'}</td>
          <td className="px-5 py-4 text-muted-foreground">{e.status?.replaceAll('_', ' ') || '—'}</td>
        </>
      )}
    />
  )
}

function displayName(value: unknown) {
  if (!value) return '—'
  if (typeof value === 'object') {
    const v = value as { name?: string }
    return v.name || '—'
  }
  return String(value) || '—'
}