"use client";

import { useState } from "react";
import Link from "next/link";

export default function RegisterPage() {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const res = await fetch("http://localhost:8080/api/register", {
            method: "POST",
            body: JSON.stringify({ email, password }),
            headers: { "Content-Type": "application/json" },
        });
        if (res.ok) {
            alert("Registration successful! Please login.");
            window.location.href = "/auth/login";
        } else {
            const data = await res.json();
            alert(data.message || "Registration failed");
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center p-6">
            <div className="card-premium w-full max-w-md">
                <div className="mb-8">
                    <h1 className="text-3xl font-bold mb-2">Join Finance AI</h1>
                    <p className="text-gray-400">Start organizing your bank movements today.</p>
                </div>

                <form onSubmit={handleSubmit} className="flex flex-col gap-5">
                    <div className="flex flex-col gap-2">
                        <label className="text-sm font-medium text-gray-400">Email Address</label>
                        <input
                            type="email"
                            placeholder="name@company.com"
                            className="input-premium"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            required
                        />
                    </div>

                    <div className="flex flex-col gap-2">
                        <label className="text-sm font-medium text-gray-400">Password</label>
                        <input
                            type="password"
                            placeholder="••••••••"
                            className="input-premium"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                        />
                    </div>

                    <button type="submit" className="btn-premium mt-4">
                        Create Account
                    </button>
                </form>

                <p className="text-center mt-8 text-sm text-gray-400">
                    Already have an account?{" "}
                    <Link href="/auth/login" className="text-p-blue font-medium hover:underline">
                        Sign in
                    </Link>
                </p>
            </div>
        </div>
    );
}
