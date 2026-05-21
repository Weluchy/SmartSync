import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { X, MessageSquare, Edit3, Eye } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import { api } from '../../api/client'; 

export default function TaskModal({ isOpen, onClose, onSave, projectId, initialData }) {
  const [formData, setFormData] = useState({
    title: '', description: '', opt: 1, real: 2, pess: 3, status: 'todo'
  });
  const [members, setMembers] = useState([]);
  const [logs, setLogs] = useState([]);
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');
  const [isPreview, setIsPreview] = useState(false); // Для Markdown
  const [activeTab, setActiveTab] = useState('details'); // details, logs, comments

  useEffect(() => {
    if (isOpen && projectId) {
      api.get(`/projects/${projectId}/members`).then(data => setMembers(data || []));
    }
    
    if (initialData) {
      setFormData({ 
        ...initialData, 
        description: initialData.description || '', 
        assignee_id: initialData.assignee_id ?? '' 
      });
      if (initialData.id) {
        // Грузим логи
        api.get(`/logs/${initialData.id}`).then(res => setLogs(res || [])).catch(() => {});
        // Грузим комментарии (ожидаем, что бэкенд отдает массив)
        api.get(`/tasks/${initialData.id}/comments`).then(res => setComments(res || [])).catch(() => {});
      }
    } else {
      setFormData({ title: '', description: '', opt: 1, real: 2, pess: 3, status: 'todo', assignee_id: '' });
      setLogs([]); 
      setComments([]);
      setActiveTab('details');
    }
  }, [initialData, isOpen, projectId]); 

  if (!isOpen) return null;

  const handleSubmit = (e) => {
    e.preventDefault();
    onSave({
      ...formData,
      assignee_id: formData.assignee_id ? parseInt(formData.assignee_id, 10) : null,
      opt: parseInt(formData.opt, 10) || 0,
      real: parseInt(formData.real, 10) || 0,
      pess: parseInt(formData.pess, 10) || 0,
      project_id: parseInt(projectId, 10) 
    });
  };

  const handlePostComment = async () => {
    if (!newComment.trim() || !initialData?.id) return;
    try {
      await api.post(`/tasks/${initialData.id}/comments`, { text: newComment });
      setNewComment('');
      // Обновляем список
      const res = await api.get(`/tasks/${initialData.id}/comments`);
      setComments(res || []);
    } catch (err) { 
      console.error(err); 
      alert("Ошибка при отправке комментария"); 
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-2xl h-[85vh] flex flex-col">
        {/* Header */}
        <div className="flex justify-between items-center p-6 border-b shrink-0">
          <h2 className="text-xl font-bold text-gray-800">
            {initialData ? `Задача ID-${initialData.id}` : 'Новая задача'}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 transition-colors"><X size={24} /></button>
        </div>

        {/* Tabs */}
        {initialData && (
          <div className="flex border-b px-6 gap-6 shrink-0 text-sm font-bold text-gray-500">
            <button onClick={() => setActiveTab('details')} className={`py-3 border-b-2 transition-colors ${activeTab === 'details' ? 'border-blue-600 text-blue-600' : 'border-transparent hover:text-gray-800'}`}>Детали</button>
            <button onClick={() => setActiveTab('comments')} className={`py-3 border-b-2 flex items-center gap-2 transition-colors ${activeTab === 'comments' ? 'border-blue-600 text-blue-600' : 'border-transparent hover:text-gray-800'}`}>
              <MessageSquare size={14} /> Комментарии ({comments.length})
            </button>
            <button onClick={() => setActiveTab('logs')} className={`py-3 border-b-2 transition-colors ${activeTab === 'logs' ? 'border-blue-600 text-blue-600' : 'border-transparent hover:text-gray-800'}`}>История</button>
          </div>
        )}

        {/* Scrollable Body */}
        <div className="flex-1 overflow-y-auto p-6 bg-gray-50/50">
          
          {activeTab === 'details' && (
            <form id="task-form" onSubmit={handleSubmit} className="space-y-5">
              <div>
                <label className="block text-xs font-bold text-gray-700 uppercase mb-1">Название</label>
                <input required className="w-full border rounded-lg p-2.5 bg-white outline-none focus:ring-2 focus:ring-blue-500 shadow-sm" value={formData.title} onChange={e => setFormData({...formData, title: e.target.value})} />
              </div>

              <div>
                <div className="flex justify-between items-end mb-1">
                  <label className="block text-xs font-bold text-gray-700 uppercase">Описание (Markdown)</label>
                  <div className="flex gap-2 text-xs font-bold">
                    <button type="button" onClick={() => setIsPreview(false)} className={`flex items-center gap-1 ${!isPreview ? 'text-blue-600' : 'text-gray-400'}`}><Edit3 size={12}/> Редактор</button>
                    <button type="button" onClick={() => setIsPreview(true)} className={`flex items-center gap-1 ${isPreview ? 'text-blue-600' : 'text-gray-400'}`}><Eye size={12}/> Просмотр</button>
                  </div>
                </div>
                {isPreview ? (
                  <div className="w-full min-h-[120px] max-h-[300px] overflow-y-auto border rounded-lg p-3 bg-gray-50 prose prose-sm">
                    <ReactMarkdown>{formData.description || '*Описание отсутствует*'}</ReactMarkdown>
                  </div>
                ) : (
                  <textarea className="w-full min-h-[120px] border rounded-lg p-2.5 bg-white outline-none focus:ring-2 focus:ring-blue-500 shadow-sm" placeholder="**Жирный** или - Список" value={formData.description} onChange={e => setFormData({...formData, description: e.target.value})} />
                )}
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="block text-[10px] font-black text-green-600 uppercase mb-1">Оптим. (ч)</label>
                  <input type="number" min="0" step="1" required className="w-full border rounded-lg p-2 bg-white shadow-sm outline-none" value={formData.opt} onChange={e => setFormData({...formData, opt: e.target.value})} />
                </div>
                <div>
                  <label className="block text-[10px] font-black text-blue-600 uppercase mb-1">Реал. (ч)</label>
                  <input type="number" min="0" step="1" required className="w-full border rounded-lg p-2 bg-white shadow-sm outline-none" value={formData.real} onChange={e => setFormData({...formData, real: e.target.value})} />
                </div>
                <div>
                  <label className="block text-[10px] font-black text-red-600 uppercase mb-1">Пессим. (ч)</label>
                  <input type="number" min="0" step="1" required className="w-full border rounded-lg p-2 bg-white shadow-sm outline-none" value={formData.pess} onChange={e => setFormData({...formData, pess: e.target.value})} />
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold text-gray-700 uppercase mb-1">Исполнитель</label>
                <select className="w-full border rounded-lg p-2.5 bg-white outline-none shadow-sm" value={formData.assignee_id || ''} onChange={e => setFormData({...formData, assignee_id: e.target.value})}>
                  <option value="">Не назначен</option>
                  {members.map(m => (<option key={m.user_id} value={m.user_id}>{m.username}</option>))}
                </select>
              </div>
            </form>
          )}

          {activeTab === 'comments' && (
            <div className="flex flex-col h-full space-y-4">
              <div className="flex-1 space-y-3">
                {comments.length > 0 ? comments.map(c => (
                  <div key={c.id} className="bg-white border rounded-xl p-3 shadow-sm">
                    <div className="flex justify-between items-center mb-1">
                      <span className="font-bold text-xs text-blue-600">{c.user_name || 'Пользователь'}</span>
                      <span className="text-[10px] text-gray-400">{new Date(c.created_at).toLocaleString('ru-RU')}</span>
                    </div>
                    <p className="text-sm text-gray-700 whitespace-pre-wrap">{c.text}</p>
                  </div>
                )) : <p className="text-sm text-gray-400 italic text-center py-6">Пока нет комментариев</p>}
              </div>
              <div className="flex gap-2">
                <input type="text" value={newComment} onChange={e=>setNewComment(e.target.value)} onKeyPress={e => e.key === 'Enter' && handlePostComment()} placeholder="Написать комментарий..." className="flex-1 border rounded-xl p-3 text-sm outline-none focus:border-blue-500" />
                <button onClick={handlePostComment} className="bg-blue-600 text-white font-bold px-4 rounded-xl hover:bg-blue-700">Отправить</button>
              </div>
            </div>
          )}

          {activeTab === 'logs' && (
            <div className="bg-white p-3 rounded-xl border shadow-sm h-full">
              {logs.length > 0 ? (
                <ul className="space-y-3 text-sm">
                  {logs.map((log, i) => (
                    <li key={i} className="border-b pb-2 last:border-0 flex justify-between">
                      <div>
                        <span className="font-bold text-blue-700">{log.action === 'updated' ? 'Обновление' : log.action === 'created' ? 'Создание' : 'Изменение статуса'}</span>
                        {log.payload?.status && <p className="text-gray-600 mt-1 text-xs">Новый статус: <span className="font-bold">{log.payload.status}</span></p>}
                      </div>
                      <span className="text-xs text-gray-400">{new Date(log.timestamp).toLocaleString('ru-RU')}</span>
                    </li>
                  ))}
                </ul>
              ) : <p className="text-gray-500 text-sm italic text-center mt-4">История пуста</p>}
            </div>
          )}

        </div>

        {/* Footer */}
        {activeTab === 'details' && (
          <div className="p-4 border-t shrink-0 flex gap-3 bg-white">
            <button type="button" onClick={onClose} className="flex-1 py-2.5 border rounded-xl font-bold hover:bg-gray-50">Отмена</button>
            <button form="task-form" type="submit" className="flex-1 py-2.5 bg-blue-600 text-white rounded-xl font-bold hover:bg-blue-700 shadow-md">
              {initialData ? 'Сохранить' : 'Создать'}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

TaskModal.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSave: PropTypes.func.isRequired,
  projectId: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
  initialData: PropTypes.object
};