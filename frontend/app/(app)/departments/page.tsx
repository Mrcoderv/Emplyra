'use client'

import CrudListPage from '@/components/crud-list-page'
import { api, type Department } from '@/lib/api'

export default function DepartmentsPage() {
  return (
    <CrudListPage<Department>
      title="Departments"
      eyebrow="Organization"
      description="Manage the teams and reporting structure behind your people operations."
      actionLabel="New department"
      empty="No departments have been created yet."
      fields={[
        { name: 'name', label: 'Department name', required: true, placeholder: 'e.g. Engineering' },
        { name: 'code', label: 'Code', required: true, placeholder: 'e.g. ENG' },
        { name: 'description', label: 'Description', type: 'textarea', span2: true, placeholder: 'What does this team own?' },
      ]}
      columns={['code', 'description', 'status']}
      load={(query) => api.departments({ search: query, page: 1, pageSize: 100 }).then((r) => r.items ?? [])}
      create={(body) => api.department.create(body)}
      update={(id, body) => api.department.update(id, body)}
      remove={(id) => api.department.remove(id)}
      rowLabel={(d) => String(d.name)}
    />
  )
}