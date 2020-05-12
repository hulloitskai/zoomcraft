/** @jsx jsx */
import { jsx, css } from "@emotion/core";
import styled from "@emotion/styled";
import { Plus } from "react-feather";

const Card = styled.div`
  width: 12rem;
  height: 12rem;
  padding: 0.9rem 0.8rem 0.8rem 0.8rem;
  margin: 0 0.8rem 0.8rem 0;
  box-sizing: border-box;

  position: relative;
  display: flex;
  flex-direction: column;
  align-items: stretch;

  h1 {
    color: white;
    margin: 0;
    font-size: 1.2rem;
    font-weight: 700;
    text-transform: uppercase;
  }
`;

export default Card;

export const AddCard = ({ onClick }) => {
  return (
    <Card
      css={css`
        position: relative;
        background: #d6d6d6;
        justify-content: flex-end;

        cursor: pointer;
        transition: background 250ms ease-in-out;
        &:hover {
          background: #bababa;
        }
      `}
      onClick={onClick}
    >
      <div
        css={css`
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;

          display: flex;
          align-items: center;
          justify-content: center;

          svg {
            --size: 2.5rem;
            width: var(--size);
            height: var(--size);
          }
        `}
      >
        <Plus color="white" />
      </div>
      <h1>VIRTUAL</h1>
    </Card>
  );
};
