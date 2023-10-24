import React, { useContext } from "react";
import { useNavigate } from "react-router-dom";

import { ThemeContext } from "@/contexts/ThemeContext";

import Input from "@/components/inputs/Input";
import Button from "@/components/buttons/Button";
import SocialButton from "@/components/buttons/SocialButton";

import Burrito from "@/assets/illustrations/Burrito";
import EyeSlashIcon from "@/assets/icons/EyeSlashIcon";
import CoverLight from "@/assets/covers/cover-light.png";
import CoverDark from "@/assets/covers/cover-dark.png";

const Login: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const navigate = useNavigate();
  return (
    <div className="flex h-screen">
      <div className="inline-block justify-center p-6 bg-nuances-white overflow-y-scroll w-[500px] min-w-[500px]">
        <div className="flex items-center justify-center">
          <div className="flex flex-col items-start justify-center gap-10">
            <div className="flex flex-col items-start gap-2">
              <Burrito height={64} width={64} />
              <span className="text-3xl font-extrabold">
                Welcome to Burrito
              </span>
            </div>
            <div className="flex flex-col items-center justify-center gap-8">
              <Input
                className="w-[300px]"
                placeholder="Your email"
                label="Email"
              />
              <Input
                className="w-full"
                placeholder="Your password"
                label="Password"
                rightIcon={<EyeSlashIcon />}
              />
              <Button className="w-full" onClick={() => navigate("/")}>
                Login
              </Button>
              <div className="flex flex-row items-center w-full gap-4">
                <hr className="w-full" />
                <span>OR</span>
                <hr className="w-full" />
              </div>
              <div className="flex flex-col items-center justify-center gap-4">
                <SocialButton variant="github" onClick={() => navigate("/")} />
                <SocialButton variant="gitlab" onClick={() => navigate("/")} />
              </div>
            </div>
            <div className="flex flex-row items-center justify-center gap-1 bg-primary-400 p-6 w-full rounded-lg">
              <span className="text-base font-normal">
                Don't have an account ?
              </span>
              <span className="text-base font-semibold">Sign up</span>
            </div>
          </div>
        </div>
      </div>
      <div className="relative flex flex-grow flex-col pt-20 px-16 bg-background-light gap-6 overflow-hidden">
        <div className="flex flex-col">
          <span className="text-5xl font-extrabold">Burrito is a TACoS</span>
          <span className="text-base font-medium">
            Monitor the status of your layers and their impacts on your project.
          </span>
        </div>
        {/* <img
          className="absolute right-0 bottom-0 -rotate-12 translate-y-20 translate-x-28 shadow-light rounded-lg h-[600px] object-cover"
          src={CoverLight}
        /> */}
      </div>
    </div>
  );
};

export default Login;
