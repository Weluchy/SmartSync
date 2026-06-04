import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite' // <-- Добавь этот импорт

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(), // <-- И обязательно вызови плагин здесь
  ],
})