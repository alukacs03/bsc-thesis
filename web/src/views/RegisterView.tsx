import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { useNavigate } from 'react-router-dom';

interface RegisterPageProps {
    onRegister: () => void;
    onNavigateToLogin: () => void;
}

const RegisterView = ({ onRegister, onNavigateToLogin }: RegisterPageProps) => {
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const navigate = useNavigate();
  onRegister = () => navigate('/dashboard');
  onNavigateToLogin = () => navigate('/login');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    alert('Registration request submitted for approval!');
    onRegister();
  };

  return (
    <div className="min-h-screen bg-slate-100 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center space-y-2">
          <div className="w-12 h-12 bg-blue-600 rounded-lg mx-auto flex items-center justify-center">
            <div className="w-6 h-6 bg-white rounded-sm"></div>
          </div>
          <CardTitle className="text-2xl text-slate-800">Request Access</CardTitle>
          <CardDescription>Submit a request to register to this control panel @Gluon</CardDescription>
          <CardDescription>Your request will be reviewed by an administrator.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="fullName">Full Name</Label>
              <Input
                id="fullName"
                type="text"
                placeholder="John Smith"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="john.smith@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                type="text"
                placeholder="john.smith"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
              />
            </div>
            <Button type="submit" className="w-full bg-blue-600 hover:bg-blue-700">
              Submit Request
            </Button>
            <div className="text-center">
              <button
                type="button"
                onClick={onNavigateToLogin}
                className="text-sm text-blue-600 hover:text-blue-700"
              >
                Back to Sign In
              </button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

export default RegisterView;
