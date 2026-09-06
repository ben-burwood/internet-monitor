// Central repo / product facts reused across components.
const REPO_OWNER = 'ben-burwood';
const REPO_NAME = 'internet-monitor';

export const REPO_URL = `https://github.com/${REPO_OWNER}/${REPO_NAME}`;
export const RELEASES_URL = `${REPO_URL}/releases`;
export const LICENSE_URL = `${REPO_URL}/blob/main/LICENSE`;
export const README_URL = `${REPO_URL}#readme`;

export const DOCKER_IMAGE = `ghcr.io/${REPO_OWNER}/${REPO_NAME}:latest`;
export const GHCR_PACKAGE_URL = `${REPO_URL}/pkgs/container/${REPO_NAME}`;

// GitHub REST endpoints — used client-side to show the latest release version.
export const API_REPO = `https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}`;
export const API_LATEST_RELEASE = `${API_REPO}/releases/latest`;

// Inspiration credited in the README.
export const INSPIRATION = [
  { name: 'MySpeed', url: 'https://github.com/gnmyt/MySpeed' },
  { name: 'speedtest-tracker', url: 'https://github.com/alexjustesen/speedtest-tracker' },
];

// GitHub Pages serves the site under a base path (see astro.config.mjs).
const BASE = import.meta.env.BASE_URL.replace(/\/$/, '');
export const asset = (path: string) => `${BASE}${path}`;
