import PropTypes from 'prop-types'; // Добавили импорт
import { LogOut, LayoutGrid, Network } from 'lucide-react';

export default function MainLayout({ children, activeView, onSwitchView, onLogout, projectName }) {
  return (
    <div className="flex-1 flex flex-col min-w-0 bg-gray-50 text-gray-900 overflow-hidden">
      <header className="h-16 bg-white border-b border-gray-200 px-6 flex items-center justify-between shadow-sm z-20">
        <h1 className="text-lg font-bold truncate text-gray-800">
          {projectName || 'Выберите проект'}
        </h1>

        <div className="flex items-center gap-4">
          <div className="flex bg-gray-100 p-1 rounded-lg border border-gray-200">
            <button 
              onClick={() => onSwitchView('graph')}
              className={`flex items-center gap-2 px-4 py-1.5 rounded-md text-sm font-bold transition-all ${
                activeView === 'graph' ? 'bg-white shadow text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              <Network size={16} /> Граф
            </button>
            <button 
              onClick={() => onSwitchView('kanban')}
              className={`flex items-center gap-2 px-4 py-1.5 rounded-md text-sm font-bold transition-all ${
                activeView === 'kanban' ? 'bg-white shadow text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              <LayoutGrid size={16} /> Канбан
            </button>
          </div>

          <button onClick={onLogout} className="p-2 text-gray-400 hover:text-red-500 transition-colors">
            <LogOut size={20} />
          </button>
        </div>
      </header>

      <main className="flex-1 overflow-hidden relative w-full">
        {children}
      </main>
    </div>
  );
}

// ПРИКАЗ: Вот эта часть уберет все твои ошибки ESLint
MainLayout.propTypes = {
  children: PropTypes.node,
  activeView: PropTypes.string.isRequired,
  onSwitchView: PropTypes.func.isRequired,
  onLogout: PropTypes.func.isRequired,
  projectName: PropTypes.string
};