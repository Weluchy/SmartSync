import { Plus, Settings2 } from 'lucide-react';

export default function Sidebar({ projects, currentProjectId, onSelectProject, onCreateProject }) {
  return (
    <aside className="w-72 bg-white border-r border-gray-200 flex flex-col shadow-sm">
      <div className="p-6 border-b border-gray-100">
        <h2 className="text-xl font-black text-blue-600 tracking-tight">SmartSync</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-1">
        <div className="flex items-center justify-between mb-4 px-2">
          <span className="text-xs font-bold text-gray-400 uppercase tracking-widest">Проекты</span>
          <button 
            onClick={() => {
              const name = prompt('Название проекта:');
              if (name) onCreateProject(name);
            }} 
            className="text-blue-600 hover:bg-blue-50 p-1 rounded transition-colors"
          >
            <Plus size={16} />
          </button>
        </div>
        
        {projects.map(p => (
          <div 
            key={p.id}
            onClick={() => onSelectProject(p.id)}
            className={`group flex items-center justify-between p-2.5 rounded-lg cursor-pointer transition-all ${
              currentProjectId === p.id ? 'bg-blue-50 text-blue-700 shadow-sm' : 'text-gray-600 hover:bg-gray-50'
            }`}
          >
            <span className="truncate text-sm font-medium">{p.name}</span>
            <Settings2 size={14} className="opacity-0 group-hover:opacity-100 text-gray-400" />
          </div>
        ))}
      </div>
    </aside>
  );
}