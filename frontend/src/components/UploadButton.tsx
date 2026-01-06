"use client";

import { useState } from "react";

export default function UploadButton() {
    const [uploading, setUploading] = useState(false);

    const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        setUploading(true);
        const formData = new FormData();
        formData.append("file", file);

        const token = localStorage.getItem("token");
        try {
            const res = await fetch("http://localhost:8080/api/upload", {
                method: "POST",
                body: formData,
                headers: {
                    "Authorization": `Bearer ${token}`
                },
            });
            if (res.ok) {
                alert("File uploaded successfully!");
            } else {
                alert("Upload failed");
            }
        } catch (err) {
            alert("Error uploading file");
        } finally {
            setUploading(false);
        }
    };

    return (
        <div className="relative">
            <input
                type="file"
                id="file-upload"
                className="hidden"
                onChange={handleFileChange}
                disabled={uploading}
            />
            <label
                htmlFor="file-upload"
                className={`btn-premium cursor-pointer flex items-center gap-2 ${uploading ? 'opacity-50 pointer-events-none' : ''}`}
            >
                <span>{uploading ? "Uploading..." : "Upload Statement"}</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 256 256">
                    <path d="M209.39,63.06,192.94,46.61A8,8,0,0,0,187.29,44.28C140.24,31.41,105.7,40.15,81.44,71.07,67.63,88.68,60.18,111.41,59.39,138.43L41.34,115.71a8,8,0,0,0-12.63,9.85l28,36a8,8,0,0,0,12.72-.11l28-36a8,8,0,1,0-12.72-9.71L74.8,128.59c1-24.16,7.69-44.52,20-60.14,21.57-27.49,53.29-34.18,97-21.7l17.59,17.59a8,8,0,1,0,11.31-11.31h0ZM232,120a8,8,0,0,0-8,8c-1,24.16-7.69,44.52-20,60.14-21.57,27.49-53.29,34.18-97,21.7l-17.59-17.59a8,8,0,0,0-11.31,11.31l16.45,16.45a8,8,0,0,0,5.65,2.33c47.05,12.87,81.59,4.13,105.85-26.79,13.81-17.61,21.26-40.34,22.05-67.36l18.05,22.72a8,8,0,0,0,12.63-9.85Z"></path>
                </svg>
            </label>
        </div>
    );
}
