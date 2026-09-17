"use client";

import { useEffect, useState } from "react";

type HealthResponse = {
	status: string;
};

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export default function HealthStatus() {
	const [message, setMessage] = useState("Checking backend health...");

	useEffect(() => {
		if (!apiURL) {
			setMessage("NEXT_PUBLIC_API_URL is not configured.");
			return;
		}

		async function checkHealth() {
			try {
				const response = await fetch(`${apiURL}/api/health`);
				if (!response.ok) {
					throw new Error(`backend returned ${response.status}`);
				}

				const data: HealthResponse = await response.json();
				setMessage(`Backend status: ${data.status}`);
			} catch (error) {
				const detail = error instanceof Error ? error.message : "unknown error";
				setMessage(`Backend unavailable: ${detail}`);
			}
		}

		void checkHealth();
	}, []);

	return <p>{message}</p>;
}
