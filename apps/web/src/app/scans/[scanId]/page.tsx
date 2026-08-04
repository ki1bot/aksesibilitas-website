import { ScanView } from "@/components/scan-view";

type ScanPageProps = {
  params: Promise<{
    scanId: string;
  }>;
};

export default async function ScanPage({ params }: ScanPageProps) {
  const { scanId } = await params;

  return <ScanView scanId={scanId} />;
}
