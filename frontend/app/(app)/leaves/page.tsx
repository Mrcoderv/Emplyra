import ModulePlaceholder from '@/components/module-placeholder'
import { navigation } from '@/lib/navigation'
export default function Page(){return <ModulePlaceholder item={navigation.find(i=>i.href==='/leaves')!}/>} 
