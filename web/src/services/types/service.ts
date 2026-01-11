export interface SystemService {
  name: string;
  description: string;
  active_state: string;
  sub_state: string;
  unit_file_state: string;
}

