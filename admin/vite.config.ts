import { defineConfig } from "vite-plus";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  lint: { options: { typeAware: true, typeCheck: true } },
  plugins: [react(), tailwindcss()],
});
