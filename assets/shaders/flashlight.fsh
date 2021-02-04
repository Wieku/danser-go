#version 330

uniform vec2 cursorPosition;
uniform float radius;
uniform float dim;

in vec2 osuPosition;

out vec4 color;

void main()
{
    float difference = length(cursorPosition - osuPosition);

    float t = pow(clamp(difference/radius, 0f, 1f), 5f);

	color = vec4(vec3(0f), mix(t, 1f, dim));
}