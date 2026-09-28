/** @type {import('next').NextConfig} */
const nextConfig = {
  // Standalone output keeps the production Docker image small (see docs/tech-stack.md).
  output: "standalone",
};

export default nextConfig;
