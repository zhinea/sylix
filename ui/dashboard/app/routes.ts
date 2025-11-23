import { type RouteConfig, index, layout, route } from "@react-router/dev/routes";

export default [
  layout("routes/dashboard.tsx", [
    index("routes/home.tsx"),
    route("servers", "routes/dashboard.servers.tsx"),
    route("logs", "routes/dashboard.logs.tsx"),
    route("stats", "routes/dashboard.stats.tsx"),
    route("accidents", "routes/dashboard.accidents.tsx"),
    route("backups", "routes/dashboard.backups.tsx"),
  ]),
] satisfies RouteConfig;
