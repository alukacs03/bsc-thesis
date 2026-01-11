import type { User } from '@/types/User';
import { useContext, useMemo, useCallback, useEffect, useState, type ReactNode } from 'react';
import { useLocalStorage } from '@/hooks/useLocalStorage';
import { createContext } from 'react';
import { authAPI } from '@/services/api/auth';

const AuthContext = createContext(null as unknown as {
    user: User | null;
    login: (user: User) => void;
    logout: () => void;
    loading: boolean;
});

export const AuthProvider = ({ children }: { children: ReactNode }) => {
    const [user, setUser] = useLocalStorage("user", null);
    const [loading, setLoading] = useState(true);

    // Validate session on mount
    useEffect(() => {
        const validateSession = async () => {
            if (user) {
                try {
                    // Check if the JWT cookie is still valid
                    const currentUser = await authAPI.getCurrentUser();
                    setUser(currentUser);
                } catch (error) {
                    // Session invalid, clear localStorage
                    console.log('Session expired or invalid, clearing user');
                    setUser(null);
                }
            }
            setLoading(false);
        };

        validateSession();
    }, []); // Only run on mount

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
            loading,
        }),
        [user, login, logout, loading]
    );
    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = () => {
    return useContext(AuthContext);
}