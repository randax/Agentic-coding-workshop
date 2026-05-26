import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Pin the workspace root to this app so Turbopack doesn't infer a stray
  // parent lockfile (e.g. one in the home directory) as the root.
  turbopack: {
    root: __dirname,
  },
};

export default nextConfig;
