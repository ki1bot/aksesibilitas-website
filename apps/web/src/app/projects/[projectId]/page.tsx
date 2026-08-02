import type { Metadata } from "next";

import { ProjectView } from "@/components/project-view";

export const metadata: Metadata = {
  title: "Detail project",
};

export default async function ProjectPage({
  params,
}: {
  params: Promise<{
    projectId: string;
  }>;
}) {
  const { projectId } = await params;

  return <ProjectView projectId={projectId} />;
}
