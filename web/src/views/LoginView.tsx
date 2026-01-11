import React, { useState } from 'react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../components/ui/card';
import { Input } from '../components/ui/input';
import { Button } from '../components/ui/button';
import { Label } from '../components/ui/label';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { authAPI } from '@/services/api/auth';
import { handleAPIError } from '@/utils/errorHandler';
import { toast } from 'sonner';

const LoginView = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();
  const { login } = useAuth();

  const onLogin = async () => {
    if (!email || !password) {
      toast.error('Please enter email and password');
      return;
    }

    try {
      await authAPI.login({ email, password });
      const user = await authAPI.getCurrentUser();
      await login(user);
      navigate('/dashboard');
      toast.success('Welcome back!');
    } catch (error) {
      console.error('Login failed', error);
      const message = handleAPIError(error, 'login');
      toast.error(message);
    }
  };
  const onNavigateToRegister = () => navigate('/register');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onLogin();
  };

  return (
    <div className="min-h-screen bg-slate-100 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center space-y-2">
          <div className="w-12 h-12 bg-blue-600 rounded-lg mx-auto flex items-center justify-center">
            <div className="w-6 h-6 bg-white rounded-sm"></div>
          </div>
          <CardTitle className="text-2xl text-slate-800">Welcome to Gluon</CardTitle>
          <CardDescription>Sign in to your infrastructure dashboard</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="admin@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <Button type="submit" className="w-full bg-blue-600 hover:bg-blue-700">
              Sign In
            </Button>
            <div className="text-center">
              <p className="text-sm text-slate-600">
                Don't have an account?{' '}
                <button
                  type="button"
                  onClick={onNavigateToRegister}
                  className="text-blue-600 hover:text-blue-700 underline"
                >
                  Request Access
                </button>
              </p>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

export default LoginView
