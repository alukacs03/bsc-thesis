import type { User } from '@/types/User';
import { useContext, useMemo, useCallback, type ReactNode } from 'react';
import { useLocalStorage } from '@/hooks/useLocalStorage';
import { createContext } from 'react';
const AuthContext = createContext(null as unknown as {
    user: User | null;
    login: (user: User) => void;
    logout: () => void;
});

export const AuthProvider = ({ children }: { children: ReactNode }) => {
    const [user, setUser] = useLocalStorage("user", null);

    const login = useCallback(async (user: User) => {
        setUser(user);
    }, [setUser]);

    const logout = useCallback(() => {
        setUser(null);
    }, [setUser]);

    const value = useMemo(
        () => ({
            user,
            login,
            logout,
        }),
        [user, login, logout]
    );
    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = () => {
    return useContext(AuthContext);
}