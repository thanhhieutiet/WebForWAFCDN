# PATCH v2.1 — Chuyển frontend sang Vite + React (áp dụng đè lên prompt v2; chỗ nào không nhắc thì giữ nguyên)

## A. Stack frontend (thay mục "Frontend" ở §1, bỏ ràng buộc "KHÔNG npm, KHÔNG bundler")

- Vite + React + TypeScript (strict) + Tailwind CSS + shadcn/ui (Radix; component được copy vào repo)
  - React Router (chế độ SPA) + TanStack Query + React Hook Form + Zod + lucide-react + sonner + Recharts (lazy, chỉ trang admin).
- Dùng bản stable mới nhất của từng thư viện, khóa bằng package-lock.json, build bằng `npm ci`.
  Không dùng Next.js/Remix/Nuxt (cần Node runtime hoặc SSR, làm mờ ranh giới origin Go mà lab đang bảo vệ).
- Test: Vitest + Testing Library cho component chính; Playwright cho e2e luồng mua hàng + chụp ảnh các trang + axe-core.
- Không dùng font/script/ảnh từ CDN ngoài. Font self-host bằng @fontsource.

## B. Cấu trúc

- Thêm `frontend/` (Vite project: src/{app,routes,features,components,lib,api}).
  Xóa `internal/adapter/web` và mọi html/template: Go không render HTML nữa.
- `frontend/src/api/` sinh từ `api/openapi.yaml` bằng openapi-typescript + openapi-fetch.
  Thêm `npm run gen:api` và một bước CI kiểm tra file sinh ra không lệch spec.
- Clean Architecture phía Go giữ nguyên. Nội dung tutorial vẫn đi qua `GET /api/tutorial/scenarios`
  (SPA gọi API, không nhúng cứng trong frontend).
- Go vẫn có SERVE_STATIC=true để chạy local không cần nginx: phục vụ `frontend/dist` kèm SPA fallback,
  trừ các path /api, /ws, /static, /contents, /openapi.json, /robots.txt, /sitemap.xml.

## C. nginx và cache (thay các phần liên quan ở §7 và §9)

- Vite build ra `/assets/*` có hash trong tên file -> `Cache-Control: public, max-age=31536000, immutable`.
- `index.html` và mọi route SPA (`try_files $uri /index.html`, chỉ cho path ngoài danh sách loại trừ ở mục B)
  -> `Cache-Control: no-cache`. /api không bao giờ fallback về index.html (404 vẫn là JSON).
- `/static/*` và `/contents/*` vẫn là file tĩnh của lab (genassets), không lẫn với `/assets/*` của Vite.
- sitemap.xml liệt kê route SPA công khai + sản phẩm active.

## D. Bảo mật frontend (thay các ý liên quan ở §9)

- React tự escape; cấm `dangerouslySetInnerHTML` (ESLint `react/no-danger` mức error). Không inline script.
- CSP: `default-src 'self'; script-src 'self'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'`.
  Thử `style-src 'self'` trước; nếu Radix cần thì thêm `'unsafe-inline'` cho style và ghi DECISIONS.md.
- Access token JWT giữ trong bộ nhớ, khôi phục bằng sessionStorage (bọc try/catch). Không dùng localStorage cho token.
  Ghi đánh đổi vào DECISIONS.md.
- CI chạy `npm audit --omit=dev` và ESLint.

## E. Docker và dev (thay §10)

- Dockerfile thêm stage `node` (node LTS alpine): `npm ci && npm run build`.
  Target `nginx` copy `frontend/dist` vào web root, giữ /srv/static cho file genassets.
- scripts/dev.ps1 chạy Vite dev server proxy `/api` và `/ws` về Go :8080. Mọi script phải quote đường dẫn có `&`.

## F. Tutorial (sửa §8)

- Bỏ yêu cầu "SSR đọc được khi tắt JS". Giữ nguyên nội dung kịch bản, runner và quy ước `runnable=false`.
- Thêm kịch bản P0:
  C17 "Hashed assets và immutable cache": so `/assets/*.js` với `/static/js/app.js`, Purge, Cache Rule theo File extension.
  C18 "Deploy bản mới nhưng CDN vẫn giữ index.html cũ" (do cache mặc định 30 phút): sửa bằng Cache Rule No Cache
  cho `/index.html` hoặc Custom Purge. Cả hai ghi "cần quan sát thực tế" ở chỗ docs chưa nói rõ.

## G. UI/UX

- Mọi giao diện làm theo skill `labshop-ui-ux` (đính kèm). Phase 3 của v2 đổi thành:
  design system + app shell -> storefront -> admin -> tutorial + Lab Console -> e2e + ảnh chụp vào docs/ui/.
- Phase 0 thêm scaffold `frontend/` (Vite, Tailwind, shadcn/ui, ESLint, Vitest) và nối vào Dockerfile + nginx.
