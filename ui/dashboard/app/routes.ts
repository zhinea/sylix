import { type RouteConfig, index, layout, route } from "@react-router/dev/routes";

export default [
  layout("routes/dashboard.tsx", [
    index("routes/home.tsx"),
    route("servers", "routes/dashboard.servers.tsx"),
    route("logs", "routes/dashboard.logs.tsx"),
  ]),
] satisfies RouteConfig;
