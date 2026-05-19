export type Keyword = {
  id: number;
  value: string;
  createdAt: string;
};

export type MatchedCertificate = {
  id: number;
  domain: string;
  issuer: string;
  notBefore: string | null;
  notAfter: string | null;
  matchedKeyword: string;
  sourceLog: string;
  createdAt: string;
};

export type MonitorStatus = {
  lastTreeSize: number;
  lastProcessedAt: string | null;
  processedLastCycle: number;
  status: string;
};
