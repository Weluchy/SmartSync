import PropTypes from 'prop-types';
import { User, LogOut, LayoutGrid, Network, BarChart3 } from 'lucide-react';

import Notifications from '../Notifications/Notifications';

export default function MainLayout({ children, activeView, onSwitchView, onLogout, projectName, tasks }) {
  return (
    <div className="flex-1 flex flex-col min-w-0 overflow-hidden" style={{ backgroundColor: 'var(--bg-page)', color: 'var(--text-primary)' }}>
      <header className="h-16 px-6 flex items-center justify-between shadow-sm z-20" style={{ backgroundColor: 'var(--bg-card)', borderBottom: '1px solid var(--border)' }}>
        <h1 className="text-lg font-bold truncate" style={{ color: 'var(--text-primary)' }}>
          {projectName || 'SmartSync'}
        </h1>

        <div className="flex items-center gap-6">
          <div className="flex p-1 rounded-lg border" style={{ backgroundColor: 'var(--kanban-bg)', borderColor: 'var(--border)' }}>
            {[
              { id: 'graph', icon: Network, label: 'Граф' },
              { id: 'kanban', icon: LayoutGrid, label: 'Канбан' },
              { id: 'dashboard', icon: BarChart3, label: 'Статистика' },
            ].map(({ id, icon: Icon, label }) => (
              <button key={id} onClick={() => onSwitchView(id)}
                className={`flex items-center gap-2 px-4 py-1.5 rounded-md text-sm font-bold transition-all ${
                  activeView === id
                    ? 'shadow text-blue-600' 
                    : 'hover:opacity-80'
                }`}
                style={{ backgroundColor: activeView === id ? 'var(--bg-card)' : 'transparent', color: activeView === id ? 'var(--text-primary)' : 'var(--text-muted)' }}
              >
                <Icon size={16} /> {label}
              </button>
            ))}
          </div>

          <Notifications tasks={tasks || []} />

          <div className="flex items-center gap-2 border-l pl-6" style={{ borderColor: 'var(--border)' }}>
            <button onClick={() => onSwitchView('profile')}
              className={`p-2 rounded-lg transition-all ${activeView === 'profile' ? 'text-blue-600' : ''}`}
              style={{ color: activeView === 'profile' ? '' : 'var(--text-muted)' }}
            >
              <User size={20} />
            </button>
            <button onClick={onLogout} className="p-2 transition-colors" style={{ color: 'var(--text-muted)' }}>
              <LogOut size={20} />
            </button>
          </div>
        </div>
      </header>

      <main className="flex-1 overflow-hidden relative w-full">
        {children}
      </main>
    </div>
  );
}

MainLayout.propTypes = {
  children: PropTypes.node,
  activeView: PropTypes.string.isRequired,
  onSwitchView: PropTypes.func.isRequired,
  onLogout: PropTypes.func.isRequired,
  projectName: PropTypes.string,
  tasks: PropTypes.array
};
