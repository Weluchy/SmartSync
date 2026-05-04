// ПРИКАЗ: Мы просто удалили импорт React, остальное без изменений
export default function UserProfile() {
  return (
    <div className="h-full w-full bg-gray-50 p-6 overflow-hidden">
      <div className="w-full h-full max-w-4xl mx-auto flex flex-col bg-white rounded-2xl shadow-xl border overflow-hidden">
        <div className="h-[72px] px-8 border-b flex items-center bg-white">
          <h3 className="font-bold text-gray-700 uppercase text-xs tracking-widest">Личный кабинет</h3>
        </div>
        <div className="p-8">
          <div className="flex items-center gap-6 mb-8">
            <div className="w-24 h-24 bg-blue-600 rounded-full flex items-center justify-center text-white text-3xl font-black shadow-lg">
              A
            </div>
            <div>
              <h2 className="text-2xl font-bold text-gray-800">Артем</h2>
              {/* Данные из твоего профиля БГУ */}
              <p className="text-gray-500 text-sm">Студент мехмата БГУ • Математик-программист</p>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-6">
            <div className="p-5 bg-gray-50 rounded-2xl border border-gray-100">
              <p className="text-[10px] font-bold text-gray-400 uppercase tracking-tighter mb-1">Стек технологий</p>
              {/* Твой актуальный стек */}
              <p className="text-sm font-bold text-gray-700">Go, PostgreSQL, Docker, React</p>
            </div>
            <div className="p-5 bg-gray-50 rounded-2xl border border-gray-100">
              <p className="text-[10px] font-bold text-gray-400 uppercase tracking-tighter mb-1">Текущий проект</p>
              <p className="text-sm font-bold text-blue-600">SmartSync.engine</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}