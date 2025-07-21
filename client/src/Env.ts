const { VITE_API_BASE_URL, ...viteConfig } = import.meta.env;

export const Env = {
    API_BASE_URL: VITE_API_BASE_URL as string,
    __vite__: viteConfig,
};

console.log(Env);
