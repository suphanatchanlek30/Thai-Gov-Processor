import Link from "next/link";

export default function Nav() {
  return (
    <nav className="flex gap-4 border-b border-gray-200 pb-4 text-sm font-medium dark:border-gray-800">
      <Link href="/" className="hover:underline">
        Photo preset
      </Link>
      <Link href="/merge-pdf" className="hover:underline">
        Merge to PDF
      </Link>
    </nav>
  );
}
