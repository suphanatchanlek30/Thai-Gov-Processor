import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Thai Gov Photo & Doc Processor",
  description: "Resize photos and merge documents to fit Thai government submission specs.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased">
        <div className="mx-auto flex max-w-3xl flex-col gap-6 p-6 sm:p-10">{children}</div>
      </body>
    </html>
  );
}
