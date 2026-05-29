interface Props {
  title: string;
}

export function Card({ title }: Props) {
  return (
    <div className="card">
      <h1>{title}</h1>
    </div>
  );
}
