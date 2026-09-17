import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "82-0",
  description: "NBA roster-building game",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
