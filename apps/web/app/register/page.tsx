'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { authApi } from '@/lib/api';
import { setAuth } from '@/lib/auth';

export default function RegisterPage() {
    const [name, setName] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const router = useRouter();

    const handleRegister = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            const res = await authApi.register({ name, email, password });
            setAuth(res.data.token, res.data.user);
            router.push('/dashboard');
        } catch (err) {
            const error = err as { response?: { data?: { error?: string } } };
            setError(error.response?.data?.error || 'Registration failed');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center p-4" style={{ background: 'var(--bg)' }}>
            <div className="w-full max-w-md">
                <div className="mb-8">
                    <h1 className="mono text-3xl font-bold" style={{ color: 'var(--accent)' }}>
                        stackd_
                    </h1>
                    <p className="mt-2 text-sm" style={{ color: 'var(--muted)' }}>
                        personal finance, simplified
                    </p>
                </div>

                <div className="rounded-lg p-6" style={{ background: 'var(--surface)', border: '1px solid var(--border)' }}>
                    <h2 className="text-lg font-medium mb-6">create account</h2>

                    {error && (
                        <div className="mb-4 p-3 rounded text-sm" style={{ background: 'rgba(255,68,68,0.1)', border: '1px solid var(--danger)', color: 'var(--danger)' }}>
                            {error}
                        </div>
                    )}

                    <form onSubmit={handleRegister} className="space-y-4">
                        <div>
                            <label className="block text-xs mb-1" style={{ color: 'var(--muted)' }}>NAME</label>
                            <input
                                type="text"
                                value={name}
                                onChange={e => setName(e.target.value)}
                                className="w-full px-3 py-2 rounded text-sm mono outline-none"
                                style={{ background: 'var(--bg)', border: '1px solid var(--border)', color: 'var(--text)' }}
                                placeholder="your name"
                                required
                            />
                        </div>
                        <div>
                            <label className="block text-xs mb-1" style={{ color: 'var(--muted)' }}>EMAIL</label>
                            <input
                                type="email"
                                value={email}
                                onChange={e => setEmail(e.target.value)}
                                className="w-full px-3 py-2 rounded text-sm mono outline-none"
                                style={{ background: 'var(--bg)', border: '1px solid var(--border)', color: 'var(--text)' }}
                                placeholder="you@example.com"
                                required
                            />
                        </div>
                        <div>
                            <label className="block text-xs mb-1" style={{ color: 'var(--muted)' }}>PASSWORD</label>
                            <input
                                type="password"
                                value={password}
                                onChange={e => setPassword(e.target.value)}
                                className="w-full px-3 py-2 rounded text-sm mono outline-none"
                                style={{ background: 'var(--bg)', border: '1px solid var(--border)', color: 'var(--text)' }}
                                placeholder="min 8 characters"
                                required
                                minLength={8}
                            />
                        </div>
                        <button
                            type="submit"
                            disabled={loading}
                            className="w-full py-2 rounded text-sm font-medium mono transition-all"
                            style={{
                                background: loading ? 'var(--border)' : 'var(--accent)',
                                color: '#0D0D0D',
                                cursor: loading ? 'not-allowed' : 'pointer',
                            }}
                        >
                            {loading ? 'creating...' : 'create_account()'}
                        </button>
                    </form>

                    <p className="mt-4 text-xs text-center" style={{ color: 'var(--muted)' }}>
                        have an account?{' '}
                        <Link href="/login" style={{ color: 'var(--accent)' }}>
                            sign in
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    );
}