export const getToken  = () => localStorage.getItem("token");
export const getTenant = () => localStorage.getItem("tenant_id");
export const getRole   = () => localStorage.getItem("role");

export const saveAuth = (token: string, tenant_id: string, role: string) => {
  localStorage.setItem("token",     token);
  localStorage.setItem("tenant_id", tenant_id);
  localStorage.setItem("role",      role);
};

export const clearAuth = () => {
  localStorage.removeItem("token");
  localStorage.removeItem("tenant_id");
  localStorage.removeItem("role");
};

export const isLoggedIn = () => !!getToken();