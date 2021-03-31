#version 330

uniform mat4 invMatrix;

in vec2 in_position;

out vec2 osuPosition;

void main()
{
    vec4 position = vec4(in_position, 0, 1);
    gl_Position = position;
    osuPosition = (invMatrix * position).xy;
}